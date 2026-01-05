package test

import (
	"fmt"
	"io/fs"
	"reflect"
	"runtime"
	"testing"

	"github.com/chaimleib/hebcalfmt/config"
)

func CheckFS(t *testing.T, name string, want, got fs.FS) {
	t.Helper()
	if (want == nil) != (got == nil) {
		t.Errorf("%s's did not match - want:\n%#v\ngot:\n%#v", name, want, got)
		return
	} else if want == nil { // implies got == nil b/c of prev check
		return
	}
	if want, ok := want.(config.FSFunc); ok {
		got, ok := got.(config.FSFunc)
		if !ok {
			t.Errorf("%s's did not match - want:\n%#v\ngot:\n%#v",
				name, want, got)
			return
		}
		CheckFSFunc(t, name, got, want)
	}
	if want, ok := want.(config.WrapFS); ok {
		got, ok := got.(config.WrapFS)
		if !ok {
			t.Errorf("%s's did not match - want:\n%#v\ngot:\n%#v",
				name, want, got)
			return
		}
		CheckWrapFS(t, name, got, want)
	}
}

func CheckFSFunc(t *testing.T, name string, got, want config.FSFunc) {
	t.Helper()
	funcPtr := func(f any) uintptr {
		return reflect.ValueOf(f).Pointer()
	}
	funcName := func(fp uintptr) string {
		return runtime.FuncForPC(fp).Name()
	}

	wantPtr := funcPtr(want)
	gotPtr := funcPtr(got)
	if wantPtr != gotPtr {
		t.Errorf("%s's did not match - want:\n%s\ngot:\n%s",
			name, funcName(wantPtr), funcName(gotPtr))
	}
}

func CheckWrapFS(t *testing.T, name string, got, want config.WrapFS) {
	t.Helper()
	for _, field := range []struct {
		Name      string
		Want, Got any
	}{
		{"BaseDir", want.BaseDir, got.BaseDir},
		{"FS", want.FS, got.FS},
	} {
		switch typedWant := field.Want.(type) {
		case fs.FS:
			typedGot := field.Got.(fs.FS)
			CheckFS(t, fmt.Sprintf("%s.%s", name, field.Name), typedWant, typedGot)

		default:
			if field.Want != field.Got {
				t.Errorf("%s.%s's did not match - want: %v, got: %v",
					name, field.Name, field.Want, field.Got)
			}
		}
	}
}
