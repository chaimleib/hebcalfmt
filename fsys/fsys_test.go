package fsys_test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/fsys"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestLocalFS(t *testing.T) {
	fileSystem, err := fsys.LocalFS()
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
	fileSystem := fsys.NewFSFunc(testFS.Open)
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

type Match struct {
	Value string
	Mode  test.WantMode
}

func str(s string) Match {
	return Match{Value: s}
}

func rgx(s string) Match {
	return Match{Value: s, Mode: test.WantRegexp}
}

func TestFSFunc_Format(t *testing.T) {
	const (
		fpath = "hi.txt"
		want  = "hi"
	)
	mapFS := fstest.MapFS{
		fpath: &fstest.MapFile{Data: []byte(want)},
	}
	cases := []struct {
		Name string
		FS   fs.FS
		Want map[string]Match
	}{
		{
			Name: "nil",
			FS:   fsys.NewFSFunc[fs.File](nil),
			Want: map[string]Match{
				"%s":  str("fsFunc[<nil>]"),
				"%v":  str("fsFunc[<nil>]"),
				"%+v": str("fsFunc[fn:<nil> id:0x0000]"),
				"%#v": str("fsys.fsFunc{fn: <nil>, id: 0x0000}"),
			},
		},
		{
			Name: "mapFS",
			FS:   fsys.NewFSFunc(mapFS.Open),
			Want: map[string]Match{
				"%s": str("fsFunc[testing/fstest.MapFS.Open-fm]"),
				"%v": str("fsFunc[testing/fstest.MapFS.Open-fm]"),
				"%+v": rgx(
					`fsFunc\[fn:testing/fstest\.MapFS\.Open-fm id:0x[0-9a-f]{4}\]`,
				),
				"%#v": rgx(
					`fsys\.fsFunc\{fn: testing/fstest\.MapFS\.Open-fm, id: 0x[0-9a-f]{4}\}`,
				),
			},
		},
		{
			Name: "os.Open",
			FS:   fsys.NewFSFunc(os.Open),
			Want: map[string]Match{
				"%s":  str("fsFunc[os.Open]"),
				"%v":  str("fsFunc[os.Open]"),
				"%+v": rgx(`fsFunc\[fn:os\.Open id:0x[0-9a-f]{4}\]`),
				"%#v": rgx(`fsys\.fsFunc\{fn: os\.Open, id: 0x[0-9a-f]{4}\}`),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			for format, want := range c.Want {
				test.CheckStringMode(
					t,
					format,
					want.Value,
					fmt.Sprintf(format, c.FS),
					want.Mode,
				)
			}
		})
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
			OpenErr: "attempted access outside of the BaseDir sub: open ../sub/hi.txt",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			wfs := fsys.WrapFS{BaseDir: c.Sub, FS: testFS}

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

func TestWrapFS_Equal(t *testing.T) {
	mapFS := fstest.MapFS{
		"hi.txt": &fstest.MapFile{Data: []byte("hi")},
	}
	osOpen := fsys.NewFSFunc(os.Open)

	cases := []struct {
		Name string
		A    fsys.WrapFS
		B    fs.FS
		Want bool
	}{
		{
			Name: "empties",
			A:    fsys.WrapFS{},
			B:    fsys.WrapFS{},
			Want: true,
		},
		{
			Name: "empty vs nil",
			A:    fsys.WrapFS{},
			B:    nil,
			Want: false,
		},
		{
			Name: "empty vs wrapped mapFS",
			A:    fsys.WrapFS{},
			B: fsys.WrapFS{
				FS:      fsys.NewFSFunc(mapFS.Open),
				BaseDir: "other/path",
			},
			Want: false,
		},
		{
			Name: "wrapped os.Opens",
			A:    fsys.WrapFS{FS: osOpen, BaseDir: "."},
			B:    fsys.WrapFS{FS: osOpen, BaseDir: "."},
			Want: true,
		},
		{
			Name: "wrapped os.Open vs os.Open",
			A:    fsys.WrapFS{FS: osOpen, BaseDir: "."},
			B:    osOpen,
			Want: false,
		},
		{
			Name: "wrapped os.Open vs empty",
			A:    fsys.WrapFS{FS: osOpen, BaseDir: "."},
			B:    fsys.WrapFS{},
			Want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := c.A.Equal(c.B)
			if got != c.Want {
				t.Errorf("want: %v, got %v", c.Want, got)
				t.Log("A", c.A)
				t.Log("B", c.B)
			}
		})
	}
}

func TestWrapFS_Format(t *testing.T) {
	mapFS := fstest.MapFS{
		"hi.txt": &fstest.MapFile{Data: []byte("hi")},
	}
	cases := []struct {
		Name string
		FS   fs.FS
		Want map[string]Match
	}{
		{
			Name: "nil",
			FS:   fsys.WrapFS{},
			Want: map[string]Match{
				"%s":  str("WrapFS[<nil> ]"),
				"%v":  str("WrapFS[<nil> ]"),
				"%+v": str("WrapFS[FS:<nil> BaseDir:]"),
				"%#v": str(`fsys.WrapFS{FS: <nil>, BaseDir: ""}`),
			},
		},
		{
			Name: "mapFS",
			FS: fsys.WrapFS{
				FS:      fsys.NewFSFunc(mapFS.Open),
				BaseDir: "other/path",
			},
			Want: map[string]Match{
				"%s": str("WrapFS[fsFunc[testing/fstest.MapFS.Open-fm] other/path]"),
				"%v": str("WrapFS[fsFunc[testing/fstest.MapFS.Open-fm] other/path]"),
				"%+v": rgx(
					`WrapFS\[FS:fsFunc\[fn:testing/fstest\.MapFS\.Open-fm id:0x[0-9a-f]{4}\] BaseDir:other/path\]`,
				),
				"%#v": rgx(
					`fsys\.WrapFS\{FS: fsys\.fsFunc\{fn: testing/fstest\.MapFS\.Open-fm, id: 0x[0-9a-f]{4}\}, BaseDir: "other/path"\}`,
				),
			},
		},
		{
			Name: "os.Open",
			FS:   fsys.WrapFS{FS: fsys.NewFSFunc(os.Open), BaseDir: "."},
			Want: map[string]Match{
				"%s": str("WrapFS[fsFunc[os.Open] .]"),
				"%v": str("WrapFS[fsFunc[os.Open] .]"),
				"%+v": rgx(
					`WrapFS\[FS:fsFunc\[fn:os\.Open id:0x[0-9a-f]{4}\] BaseDir:\.\]`,
				),
				"%#v": rgx(
					`fsys\.WrapFS\{FS: fsys\.fsFunc\{fn: os\.Open, id: 0x[0-9a-f]{4}\}, BaseDir: "\."\}`,
				),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			for format, want := range c.Want {
				test.CheckStringMode(
					t,
					format,
					want.Value,
					fmt.Sprintf(format, c.FS),
					want.Mode,
				)
			}
		})
	}
}

func TestFSEqual(t *testing.T) {
	mapFS := fstest.MapFS{
		"hi.txt": &fstest.MapFile{Data: []byte("hi")},
	}
	var emptyWrapFS fsys.WrapFS
	wrapMap := fsys.WrapFS{FS: mapFS, BaseDir: "other/path"}
	wrapMap2 := fsys.WrapFS{FS: mapFS, BaseDir: "other/path"}
	osOpen := fsys.NewFSFunc(os.Open)
	osOpen2 := fsys.NewFSFunc(os.Open)
	wrapOpen := fsys.WrapFS{FS: osOpen, BaseDir: "."}
	wrapOpen2 := fsys.WrapFS{FS: osOpen2, BaseDir: "."}
	dirFS := os.DirFS("/usr")
	dirFS2 := os.DirFS("/usr")
	sub, err := fs.Sub(dirFS, "bin")
	if err != nil {
		t.Error("init sub:", err)
	}
	sub2, err := fs.Sub(dirFS, "bin")
	if err != nil {
		t.Error("init sub2:", err)
	}

	cases := []struct {
		Name string
		A, B fs.FS
		Want bool
	}{
		{Name: "nils", Want: true},
		{Name: "osOpens", A: osOpen, B: osOpen, Want: true},
		// TODO: find a way to fix this?
		{Name: "osOpen vs osOpen2", A: osOpen, B: osOpen2, Want: false},
		// TODO: find a way to fix this?
		{Name: "wrapOpen vs wrapOpen2", A: wrapOpen, B: wrapOpen2, Want: false},
		{Name: "osOpen vs nil", A: osOpen, B: nil, Want: false},
		{Name: "osOpen vs empty wrapFS", A: osOpen, B: emptyWrapFS, Want: false},
		{Name: "mapFSes", A: mapFS, B: mapFS, Want: true},
		{Name: "wrapFSes", A: wrapMap, B: wrapMap, Want: true},
		{Name: "wrapMap vs wrapMap2", A: wrapMap, B: wrapMap2, Want: true},
		{Name: "wrapOpens", A: wrapOpen, B: wrapOpen, Want: true},
		{Name: "dirFSes", A: dirFS, B: dirFS, Want: true},
		{Name: "dirFS vs dirFS2", A: dirFS, B: dirFS2, Want: true},
		{Name: "subs", A: sub, B: sub, Want: true},
		// TODO: fix this?
		{Name: "sub vs sub2", A: sub, B: sub2, Want: false},
		{Name: "mapFS vs osOpen", A: mapFS, B: osOpen, Want: false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := fsys.Equal(c.A, c.B)
			if c.Want != got {
				t.Errorf("want: %v, got %v", c.Want, got)
				t.Log("A", c.A)
				t.Log("B", c.B)
			}
		})
	}
}
