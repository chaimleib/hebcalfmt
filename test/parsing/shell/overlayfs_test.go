package shell_test

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing/shell"
)

type errorFS struct{}

func (errorFS) Open(fpath string) (fs.File, error) {
	if fpath == "unreadable" {
		return nil, errors.New("errorFS: error opening file")
	}
	return nil, os.ErrNotExist
}

func TestOverlayFS(t *testing.T) {
	files0 := fstest.MapFS{
		"example.txt": &fstest.MapFile{Data: []byte("hello")},
	}
	files1 := fstest.MapFS{
		"example.txt": &fstest.MapFile{Data: []byte("overridden")},
		"base.txt":    &fstest.MapFile{Data: []byte("base")},
	}
	o := shell.OverlayFS{files0, files1, errorFS{}}

	cases := []struct {
		Fpath string
		Want  string
		Err   string
	}{
		{Fpath: "example.txt", Want: "hello"},
		{Fpath: "base.txt", Want: "base"},
		{
			Fpath: "does-not-exist",
			Err:   "OverlayFS.Open does-not-exist: file does not exist",
		},
		{
			Fpath: "unreadable",
			Err:   "OverlayFS[2].Open unreadable: errorFS: error opening file",
		},
	}
	for _, c := range cases {
		t.Run(c.Fpath, func(t *testing.T) {
			got, err := fs.ReadFile(o, c.Fpath)
			test.CheckErr(t, err, c.Err)
			test.CheckString(t, "file", c.Want, string(got))
		})
	}
}
