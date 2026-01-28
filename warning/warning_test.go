package warning_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/warning"
)

func TestBuild(t *testing.T) {
	cases := []struct {
		Name  string
		Input warning.Warnings
		Err   string
	}{
		{Name: "nil"},
		{Name: "empty", Input: warning.Warnings{}},
		{
			Name:  "one",
			Input: warning.Warnings{errors.New("test error")},
			Err:   "warn: test error",
		},
		{
			Name: "two",
			Input: warning.Warnings{
				errors.New("test error one"),
				errors.New("test error two"),
			},
			Err: "2 warnings:\ntest error one\ntest error two",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.Input.Build()
			test.CheckErr(t, err, c.Err)
			if err != nil {
				if !errors.Is(err, warning.ErrWarn) {
					t.Error("expected result to be ErrWarn")
				}
			}
		})
	}
}

type MultiError interface {
	error
	Unwrap() []error
}

func TestJoin(t *testing.T) {
	testErr := errors.New("test error")
	warn1 := warning.Warnings{errors.New("test warn")}
	warn2 := warning.Warnings{
		errors.New("test warn 1"),
		errors.New("test warn 2"),
	}

	cases := []struct {
		Name     string
		Input    warning.Warnings
		Joiner   error
		Err      string
		SubWarns []string
	}{
		{Name: "nil"},
		{Name: "empty", Input: warning.Warnings{}},
		{
			Name:   "nil + err",
			Joiner: testErr,
			Err:    "test error",
		},
		{
			Name:   "empty + err",
			Input:  warning.Warnings{},
			Joiner: testErr,
			Err:    "test error",
		},
		{
			Name:     "warn1",
			Input:    warn1,
			Err:      "warn: test warn",
			SubWarns: []string{"test warn"},
		},
		{
			Name:     "warn2",
			Input:    warn2,
			Err:      "2 warnings:\ntest warn 1\ntest warn 2",
			SubWarns: []string{"test warn 1", "test warn 2"},
		},
		{
			Name:     "warn1 + err",
			Input:    warn1,
			Joiner:   testErr,
			Err:      "warn: test warn\n\nerror: test error",
			SubWarns: []string{"test warn"},
		},
		{
			Name:     "warn2 + err",
			Input:    warn2,
			Joiner:   testErr,
			Err:      "2 warnings:\ntest warn 1\ntest warn 2\n\nerror: test error",
			SubWarns: []string{"test warn 1", "test warn 2"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := c.Input.Join(c.Joiner)

			test.CheckErr(t, got, c.Err)

			// Check the expected components.

			// Got nil?
			if got == nil {
				if len(c.SubWarns) > 0 {
					t.Errorf(
						"expected %d subwarns - got nil, want:\n%v",
						len(c.SubWarns),
						strings.Join(c.SubWarns, "\n"),
					)
				}
				if c.Joiner != nil {
					t.Errorf(
						"expected the Joiner in the result - got nil, want:\n%v",
						c.Joiner,
					)
				}
				return
			}

			// Unexpected got
			if len(c.SubWarns) == 0 && c.Joiner == nil {
				t.Fatalf("want nil, got:\n%v", got)
			}

			// Just the Joiner
			if len(c.SubWarns) == 0 && c.Joiner != nil {
				if c.Joiner != got {
					t.Errorf(
						"expected the Joiner back - want:\n%v\ngot:\n%v",
						c.Joiner,
						got,
					)
				}
				return
			}

			// Something with Warnings...
			if !errors.Is(got, warning.ErrWarn) {
				t.Errorf("expected ErrWarn, got %v", got)
			}

			// Just the Warnings
			var warns []error
			if len(c.SubWarns) > 0 && c.Joiner == nil {
				pairMulti, ok := got.(MultiError)
				if !ok {
					t.Fatalf(
						"expected pair length 2, got simple error - got:\n%v",
						got)
				}

				pair := pairMulti.Unwrap()
				if len(pair) != 2 {
					t.Fatalf(
						"expected pair length 2, got %d:\n%v",
						len(pair), strings.Join(test.AsStrings(pair), "\n"))
				}

				if pair[0] != warning.ErrWarn {
					t.Errorf(
						"expected pair[0] to be ErrWarn, got:\n%v",
						pair[0],
					)
				}

				multi, ok := pair[1].(MultiError)
				if !ok {
					if len(c.SubWarns) > 1 {
						t.Errorf(
							"expected pair[1] unwrappable, got simple error - got:\n%v",
							pair[1])
					}
					warns = warning.Warnings{pair[1]}
				} else {
					warns = multi.Unwrap()
				}
			} else {
				// Combo: Warnings + Joiner
				multi, ok := got.(MultiError)
				if !ok {
					t.Fatalf(
						"expected 2 unwrapped parts, got simple error:\n%v",
						got,
					)
				}

				pair := multi.Unwrap()
				if len(pair) != 2 {
					t.Fatalf(
						"expected 2 unwrapped parts, got %d parts - got:\n%#v",
						len(pair),
						pair,
					)
				}

				if c.Joiner != pair[1] {
					t.Errorf(
						"expected the Joiner as the error part - want:\n%v\ngot:\n%v",
						c.Joiner, pair[1],
					)
				}

				// extract the subwarns
				warnPairMulti, ok := pair[0].(MultiError)
				if !ok {
					t.Fatalf("expected a warnPair, got:\n%v", pair[0])
				}

				warnPair := warnPairMulti.Unwrap()
				if len(warnPair) != 2 {
					t.Fatalf(
						"expected a warnPair len 2, got %d:\n%v",
						len(warnPair),
						strings.Join(test.AsStrings(warnPair), "\n"),
					)
				}
				if warnPair[0] != warning.ErrWarn {
					t.Errorf(
						"expected warnPair[0] to be ErrWarn - got:\n%v",
						warnPair[0],
					)
				}

				multiSubs, ok := warnPair[1].(MultiError)
				if !ok {
					if len(c.SubWarns) > 1 {
						t.Errorf(
							"expected warnPair[1] to be unwrappable, got a simple err:\n%v",
							warnPair[1],
						)
					}
					warns = warning.Warnings{warnPair[1]}
				} else {
					warns = multiSubs.Unwrap()
				}
			}

			// check the subwarns
			for i, warn := range warns {
				if i >= len(c.SubWarns) {
					t.Fatalf(
						"got %d subwarns, expected %d - got:\n%v\ngot extra:\n%s",
						len(warns),
						len(c.SubWarns),
						strings.Join(test.AsStrings(warns), "\n"),
						strings.Join(test.AsStrings(warns[i:]), "\n"),
					)
				}

				test.CheckString(
					t,
					fmt.Sprintf("subwarn[%d]", i),
					c.SubWarns[i],
					warn.Error(),
					test.WantEqual,
				)
			}

			if len(c.SubWarns) > len(warns) {
				t.Fatalf(
					"got %d subwarns, expected %d - missing:\n%#v",
					len(warns),
					len(c.SubWarns),
					strings.Join(c.SubWarns[len(warns):], "\n"),
				)
			}
		})
	}
}

func TestAppend(t *testing.T) {
	testErr := errors.New("test error")
	cases := []struct {
		Name     string
		Warnings warning.Warnings
		Joiner   error
		Want     warning.Warnings
	}{
		{Name: "nil + nil"},
		{Name: "empty + nil", Warnings: warning.Warnings{}},
		{Name: "nil + err", Joiner: testErr, Want: warning.Warnings{testErr}},
		{
			Name:     "empty + err",
			Warnings: warning.Warnings{},
			Joiner:   testErr,
			Want:     warning.Warnings{testErr},
		},
		{
			Name:     "warn1 + err",
			Warnings: warning.Warnings{testErr},
			Joiner:   testErr,
			Want:     warning.Warnings{testErr, testErr},
		},
		{
			Name:     "warn1 + nil",
			Warnings: warning.Warnings{testErr},
			Want:     warning.Warnings{testErr},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := c.Warnings
			got.Append(c.Joiner)
			test.CheckSlice(t, "warnings", c.Want, got)
		})
	}
}
