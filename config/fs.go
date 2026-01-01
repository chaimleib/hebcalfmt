package config

import (
	"io/fs"
	"os"
)

// DefaultFS returns a new [fs.FS] for use in loading configuration files.
var DefaultFS = LocalFS

// LocalFS returns a new [fs.FS] that relays `Open()` calls to [os.Open].
// Unlike [os.DirFS], this proxy can return a wrapped [os.ErrNotExist].
func LocalFS() (fs.FS, error) { return NewFSFunc(os.Open), nil }

// FSFunc turns a function that returns [fs.File]s into an [fs.FS].
type FSFunc func(string) (fs.File, error)

var _ fs.FS = FSFunc(nil)

func (fsf FSFunc) Open(fpath string) (fs.File, error) {
	return fsf(fpath)
}

// NewFSFunc turns a function that returns a type compatible with [fs.File]
// into an [fs.FS].
func NewFSFunc[F fs.File](fn func(string) (F, error)) FSFunc {
	open := func(fpath string) (f fs.File, err error) {
		f, err = fn(fpath)
		return f, err
	}
	return FSFunc(open)
}
