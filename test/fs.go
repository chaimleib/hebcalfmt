package test

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"reflect"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/config"
)

func CheckFS(t Test, name string, want, got fs.FS) {
	t.Helper()
	if (want == nil) != (got == nil) {
		t.Errorf("%s's did not match - want:\n%v\ngot:\n%v", name, want, got)
		return
	} else if want == nil { // implies got == nil b/c of prev check
		return
	}
	switch typedWant := want.(type) {
	case config.FSFunc:
		CheckFSMatchThen(t, name, CheckFSFunc, typedWant, got)

	case config.WrapFS:
		CheckFSMatchThen(t, name, CheckWrapFS, typedWant, got)

	case fstest.MapFS:
		CheckFSMatchThen(t, name, CheckMapFS, typedWant, got)

	default:
		t.Errorf(
			"unknown fs.FS types:\nwant is a %T\ngot  is a %T\nwant:\n%v\ngot:\n%v",
			want, got, want, got,
		)
	}
}

func CheckFSMatchThen[T any](
	t Test,
	name string,
	checker CheckerFunc[T],
	want T,
	got any,
) {
	typedGot, ok := got.(T)
	if !ok {
		t.Errorf(
			"%s's types did not match -\n  want %T:\n%v\n  got  %T:\n%v",
			name,
			want, want,
			got, got,
		)
		return
	}
	checker(t, name, want, typedGot)
}

func CheckFSFunc(t Test, name string, want, got config.FSFunc) {
	t.Helper()
	if !want.Equal(got) {
		t.Errorf("%s's did not match - want:\n%v\ngot:\n%v",
			name, want, got)
	}
}

func CheckWrapFS(t Test, name string, want, got config.WrapFS) {
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

func CheckMapFS(t Test, name string, want, got fstest.MapFS) {
	t.Helper()
	mapFilesEqual := func(a, b *fstest.MapFile) bool {
		if a == nil && b == nil {
			return true
		} else if a == nil || b == nil {
			return false
		}
		return bytes.Equal(a.Data, b.Data) && a.Mode == b.Mode
	}

	if !maps.EqualFunc(want, got, mapFilesEqual) {
		t.Errorf("%s's did not match - want:\n%#v\ngot:\n%#v", name, want, got)
	}
}

func FSEqual(a, b fs.FS) bool {
	switch a := a.(type) {
	case interface{ Equal(any) bool }:
		return a.Equal(b)

	case fstest.MapFS:
		b, ok := b.(fstest.MapFS)
		if !ok {
			return false
		}
		return maps.Equal(a, b)

	default:
		if fmt.Sprintf("%#v", a) == fmt.Sprintf("%#v", b) {
			aPtr := reflect.ValueOf(a.Open).Pointer()
			bPtr := reflect.ValueOf(b.Open).Pointer()
			return aPtr == bPtr
		}
		return false
	}
}
