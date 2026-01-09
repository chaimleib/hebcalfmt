package test

import (
	"io/fs"

	"github.com/chaimleib/hebcalfmt/fsys"
)

func CheckFS(t Test, name string, want, got fs.FS) {
	t.Helper()
	if (want == nil) != (got == nil) {
		t.Errorf(
			"%s's nilness did not match - want:\n%v\ngot:\n%v",
			name, want, got,
		)
		return
	} else if want == nil { // implies got == nil b/c of prev check
		return
	}

	if !fsys.Equal(want, got) {
		t.Errorf(
			"%s's did not match -\n  want %T:\n%v\n  got  %T:\n%v",
			name,
			want, want,
			got, got,
		)
	}
}
