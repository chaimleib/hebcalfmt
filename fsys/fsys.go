package fsys

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sync/atomic"
)

// DefaultFS returns a new [fs.FS] for accessing the local filesystem.
var DefaultFS = LocalFS

// LocalFS returns a new [fs.FS] that relays `Open()` calls to [os.Open].
// Unlike [os.DirFS], this proxy can return a wrapped [os.ErrNotExist].
func LocalFS() (fs.FS, error) { return NewFSFunc(os.Open), nil }

// fsFunc turns a function that returns [fs.File]s into an [fs.FS].
type fsFunc struct {
	fn func(string) (fs.File, error)

	// name is displayed when formatting an fsFunc.
	// This must be saved before it gets wrapped in a closure.
	name string

	// id is for checking equality by unique ID.
	id uint64
}

var fsFuncLastID atomic.Uint64

var _ fs.FS = (*fsFunc)(nil)

// NewFSFunc turns a function that returns a type compatible with [fs.File]
// into an [fs.FS].
func NewFSFunc[F fs.File](fn func(string) (F, error)) fs.FS {
	if fn == nil {
		// All nil fsFuncs compare equal.
		// Empty values can be useful for the type info.
		return fsFunc{}
	}

	funcPtr := reflect.ValueOf(fn).Pointer()
	funcName := runtime.FuncForPC(funcPtr).Name()
	return fsFunc{
		fn: func(fpath string) (fs.File, error) {
			return fn(fpath)
		},
		name: funcName,
		id:   fsFuncLastID.Add(1),
	}
}

func (fsf fsFunc) Open(fpath string) (fs.File, error) {
	return fsf.fn(fpath)
}

// Equal returns whether other is the same instance of fsFunc.
// Every invocation of NewFSFunc produces an instance with a new id.
func (fsf fsFunc) Equal(other any) bool {
	typedOther, ok := other.(fsFunc)
	if !ok {
		return false
	}
	return fsf.id == typedOther.id
}

func (fsf fsFunc) Format(state fmt.State, verb rune) {
	name := fsf.name
	if fsf.id == 0 {
		name = "<nil>"
	}
	vFormats := map[rune]string{
		'+': "fsFunc[fn:%s id:0x%04x]",
		'#': "fsys.fsFunc{fn: %v, id: 0x%04x}",
	}
	format := "fsFunc[%s]"
	if verb == 'v' {
		for r, specialForm := range vFormats {
			if state.Flag(int(r)) {
				format = specialForm
				fmt.Fprintf(state, format, name, fsf.id)
				return
			}
		}
	}
	fmt.Fprintf(state, format, name)
}

type WrapFS struct {
	FS      fs.FS
	BaseDir string
}

// Equal returns true if the types are the same,
// the BaseDirs are the same, and the FS fields are equal under FSEqual.
func (w WrapFS) Equal(other any) bool {
	typedOther, ok := other.(WrapFS)
	if !ok {
		return false
	}
	if w.BaseDir != typedOther.BaseDir {
		return false
	}

	return Equal(w.FS, typedOther.FS)
}

func (w WrapFS) Open(fpath string) (fs.File, error) {
	if !filepath.IsLocal(fpath) {
		return nil, fmt.Errorf(
			"attempted access outside of the BaseDir %s: open %s",
			w.BaseDir,
			fpath,
		)
	}
	// use path instead of filepath, which is os-sensitive.
	// os.Open will work on Windows
	// with forward slashes instead of its preferred backslashes;
	// but internet-based FSes and so on are likely to break if given backslashes.
	resolved := path.Join(w.BaseDir, fpath)
	return w.FS.Open(resolved)
}

func (w WrapFS) Format(state fmt.State, verb rune) {
	vFormats := map[rune]string{
		'+': "WrapFS[FS:%+v BaseDir:%s]",
		'#': "fsys.WrapFS{FS: %#v, BaseDir: %q}",
	}
	format := "WrapFS[%v %s]"
	if verb == 'v' {
		for r, specialForm := range vFormats {
			if state.Flag(int(r)) {
				format = specialForm
				break
			}
		}
	}
	fmt.Fprintf(state, format, w.FS, w.BaseDir)
}

func Equal(a, b fs.FS) bool {
	if a == nil && b == nil {
		return true
	} else if (a == nil) || (b == nil) {
		return false
	}

	type Equaler interface{ Equal(any) bool }

	switch a := a.(type) {
	case Equaler:
		return a.Equal(b)

	// case fstest.MapFS:
	// 	b, ok := b.(fstest.MapFS)
	// 	return ok && maps.Equal(a, b)
	//
	default:
		aVal := reflect.ValueOf(a)
		bVal := reflect.ValueOf(b)
		if aVal.Comparable() && bVal.Comparable() {
			return a == b
		}
		aType := aVal.Type()
		if aType != bVal.Type() {
			return false
		}

		return reflect.DeepEqual(a, b)
	}
}
