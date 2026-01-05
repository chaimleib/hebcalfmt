package config_test

import (
	"io"
	"testing"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/config"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestLocalFS(t *testing.T) {
	fileSystem, err := config.LocalFS()
	if err != nil {
		t.Errorf("failed to load LocalFS: %v", err)
		return
	}

	const fpath = "testdata/ok.txt"
	f, err := fileSystem.Open(fpath)
	if err != nil {
		t.Errorf("failed to open %s: %v", fpath, err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		t.Errorf("failed to read: %v", err)
	}

	const want = "ok\n"
	if string(buf) != "ok\n" {
		t.Errorf("want: %q, got: %q", want, string(buf))
	}
}

func TestFSFunc(t *testing.T) {
	const (
		fpath = "hi.txt"
		want  = "hi"
	)
	testFS := fstest.MapFS{
		fpath: &fstest.MapFile{Data: []byte(want)},
	}
	fileSystem := config.NewFSFunc(testFS.Open)
	f, err := fileSystem.Open(fpath)
	if err != nil {
		t.Errorf("failed to open %s: %v", fpath, err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		t.Errorf("failed to read: %v", err)
	}

	if string(buf) != want {
		t.Errorf("want: %q, got: %q", want, string(buf))
	}
}

func TestWrapFS(t *testing.T) {
	const want = "hi, I'm Sub!"
	testFS := fstest.MapFS{
		"hi.txt":     &fstest.MapFile{Data: []byte("hi, I'm Root!")},
		"sub/hi.txt": &fstest.MapFile{Data: []byte(want)},
	}
	cases := []struct {
		Name    string
		Sub     string
		Fpath   string
		Want    string
		OpenErr string
	}{
		{Name: "sub", Sub: "sub", Fpath: "hi.txt", Want: "hi, I'm Sub!"},
		{
			Name:    "double dots",
			Sub:     "sub",
			Fpath:   "../sub/hi.txt",
			OpenErr: "attempted access outside of the config file's directory tree: open ../sub/hi.txt",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			wfs := config.WrapFS{BaseDir: c.Sub, FS: testFS}

			f, err := wfs.Open(c.Fpath)
			test.CheckErr(t, err, c.OpenErr)
			if err != nil {
				return
			}
			defer f.Close()

			buf, err := io.ReadAll(f)
			if err != nil {
				t.Errorf("unexpected error while reading %s from %s: %v",
					c.Fpath, c.Sub, err)
			}

			if want != string(buf) {
				t.Errorf("contents of %s from %s did not match - want:\n%s\ngot:\n%s",
					c.Fpath, c.Sub, want, string(buf))
			}
		})
	}
}
