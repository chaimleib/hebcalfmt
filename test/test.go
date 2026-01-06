package test

import (
	"testing"
)

func CheckErr(t *testing.T, got error, wantErr string) {
	t.Helper()

	if wantErr == "" {
		if got != nil {
			t.Errorf("want nil err, got %q", got.Error())
		}
		return
	}

	if got == nil {
		t.Errorf("got nil err, want %q", wantErr)
		return
	}
	if got.Error() != wantErr {
		t.Errorf("got wrong err:\n%q\nwant:\n%q", got.Error(), wantErr)
	}
}

// CheckNilPtrThen returns true only if both want and got are not nil.
// It fails the test only if the nilness is inconsistent.
// This is designed to simplify boilerplate in equality checks,
// so that type-specific equality checkers can skip casting from any
// and checking for nil. For example:
//
//	CheckNilPtrThen(t, CheckDateRange, field.Name, typedWant, field.Got)
func CheckNilPtrThen[T any](
	t *testing.T,
	checker func(t *testing.T, name string, want, got T),
	name string,
	typedWant *T,
	got any,
) {
	t.Helper()
	typedGot, ok := got.(*T)
	if !ok {
		t.Errorf("%s's types did not match - want:\n%T\ngot:\n%T",
			name, typedWant, got)
	} else if (typedWant == nil) != (typedGot == nil) {
		t.Errorf("%s's did not match - want:\n%#v\ngot:\n%#v",
			name, typedWant, typedGot)
	} else if typedWant != nil { // implies got != nil b/c of prev check
		checker(t, name, *typedWant, *typedGot)
	}
}

func CheckComparable[T comparable](
	t *testing.T,
	name string,
	want, got T,
) {
	t.Helper()
	if want != got {
		t.Errorf("%s's did not match - want:\n%#v\ngot:\n%#v",
			name, want, got)
	}
}
