package test_test

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/fsys"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestCheckFS(t *testing.T) {
	const fieldName = "FS"
	fsFunc := fsys.NewFSFunc(os.Open)
	osFS := os.DirFS("some/path")
	mapFS := fstest.MapFS{"stub.txt": nil}
	wrapFS := fsys.WrapFS{
		BaseDir: "other/path",
		FS:      fsFunc,
	}
	wrapFSCopy := wrapFS

	cases := []struct {
		Name      string
		WantInput fs.FS
		GotInput  fs.FS
		Want      string
	}{
		{Name: "nils"},
		{
			Name:      "os.DirFS vs nil",
			WantInput: osFS,
			GotInput:  nil,
			Want: `FS's nilness did not match - want:
some/path
got:
<nil>
`,
		},
		{
			Name:      "nil vs mapFS",
			WantInput: nil,
			GotInput:  mapFS,
			Want: `FS's nilness did not match - want:
<nil>
got:
map[stub.txt:<nil>]
`,
		},
		{
			Name:      "fsFunc vs mapFS",
			WantInput: fsFunc,
			GotInput:  mapFS,
			Want: `FS's did not match -
  want fsys.fsFunc:
fsFunc[os.Open]
  got  fstest.MapFS:
map[stub.txt:<nil>]
`,
		},
		{
			Name:      "wrapFS vs nil",
			WantInput: wrapFS,
			GotInput:  nil,
			Want: `FS's nilness did not match - want:
WrapFS[fsFunc[os.Open] other/path]
got:
<nil>
`,
		},
		{
			Name:      "wrapFS vs wrapFSCopy",
			WantInput: wrapFS,
			GotInput:  wrapFSCopy,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)

			test.CheckFS(mockT, fieldName, c.WantInput, c.GotInput)

			if wantFailed := c.Want != ""; wantFailed != mockT.Failed() {
				t.Errorf("wantFailed is %v, but t.Failed() is %v",
					wantFailed, mockT.Failed())
			}

			if gotLogs := mockT.buf.String(); c.Want != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Want, gotLogs)
			}
		})
	}
}
