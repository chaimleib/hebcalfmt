package config

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
)

// DefaultFS returns a new [fs.FS] for use in loading configuration files.
var DefaultFS = LocalFS

// LocalFS returns a new [fs.FS] that relays `Open()` calls to [os.Open].
// Unlike [os.DirFS], this proxy can return a wrapped [os.ErrNotExist].
func LocalFS() (fs.FS, error) { return NewFSFunc(os.Open), nil }

// FSFunc turns a function that returns [fs.File]s into an [fs.FS].
type FSFunc struct {
	fn   func(string) (fs.File, error)
	name string
}

var _ fs.FS = (*FSFunc)(nil)

func (fsf FSFunc) Open(fpath string) (fs.File, error) {
	return fsf.fn(fpath)
}

func (fsf FSFunc) Format(state fmt.State, verb rune) {
	name := fsf.name
	if fsf.name == "" {
		name = "<nil>"
	}
	fmt.Fprintf(state, "FSFunc[%s]", name)
}

// NewFSFunc turns a function that returns a type compatible with [fs.File]
// into an [fs.FS].
func NewFSFunc[F fs.File](fn func(string) (F, error)) fs.FS {
	funcPtr := reflect.ValueOf(fn).Pointer()
	funcName := runtime.FuncForPC(funcPtr).Name()
	return FSFunc{
		fn: func(fpath string) (fs.File, error) {
			return fn(fpath)
		},
		name: funcName,
	}
}

type WrapFS struct {
	BaseDir string
	FS      fs.FS
}

func (w WrapFS) Open(fpath string) (fs.File, error) {
	if !filepath.IsLocal(fpath) {
		return nil, fmt.Errorf(
			"attempted access outside of the config file's directory tree: open %s",
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
