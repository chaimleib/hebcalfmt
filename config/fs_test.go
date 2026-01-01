package config_test

import (
	"io"
	"testing"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/config"
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
