package test

import (
	"testing"
)

// Test is a mock for testing.T.
type Test interface {
	Errorf(string, ...any)
	Log(...any)
	Fail()
	Failed() bool
	Helper()
	Cleanup(func())
}

var _ Test = (*testing.T)(nil)

func CheckErr(t Test, got error, wantErr string) {
	t.Helper()

	if wantErr == "" {
		if got != nil {
			t.Errorf("want nil err, got:\n%s", got.Error())
		}
		return
	}

	if got == nil {
		t.Errorf("got nil err, want:\n%s", wantErr)
		return
	}
	if got.Error() != wantErr {
		t.Errorf("got wrong err:\n%v\nwant:\n%s", got, wantErr)
	}
}

// CheckerFunc is a type for generic test helper functions
// which compare want with got.
type CheckerFunc[T any] func(t Test, name string, want, got T)

// CheckNilPtrThen returns true only if both want and got are not nil.
// It fails the test only if the nilness is inconsistent.
// This is designed to simplify boilerplate in equality checks,
// so that type-specific equality checkers can skip casting from any
// and checking for nil. For example:
//
//	CheckNilPtrThen(t, CheckDateRange, field.Name, typedWant, field.Got)
func CheckNilPtrThen[T any](
	t Test,
	checker CheckerFunc[T],
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
		if typedWant == nil {
			t.Errorf("%s's did not match - want:\n(%T)(nil)\ngot pointer to:\n%#v",
				name, typedWant, *typedGot)
		} else {
			t.Errorf("%s's did not match - want pointer to:\n%#v\ngot:\n(%T)(nil)",
				name, *typedWant, typedGot)
		}
	} else if typedWant != nil { // implies got != nil b/c of prev check
		checker(t, name, *typedWant, *typedGot)
	}
}

var _ CheckerFunc[int] = CheckComparable[int]

func CheckComparable[T comparable](
	t Test,
	name string,
	want, got T,
) {
	t.Helper()
	if want != got {
		t.Errorf("%s's did not match - want:\n%#v\ngot:\n%#v",
			name, want, got)
	}
}
