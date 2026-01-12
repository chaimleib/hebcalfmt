package test_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
)

func TestAsString(t *testing.T) {
	type KV struct {
		Key   string
		Value string
	}
	cases := []struct {
		Name  string
		Input any
		Want  []string
	}{
		{Name: "empty ints", Input: []int(nil), Want: nil},
		{Name: "ints", Input: []int{1, 2, 3}, Want: []string{"1", "2", "3"}},
		{
			Name:  "strings",
			Input: strings.Fields("hi bye"),
			Want:  strings.Fields("hi bye"),
		},
		{
			Name:  "KVs",
			Input: []KV{{Key: "myKey", Value: "myValue"}},
			Want:  []string{"{myKey myValue}"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var got []string
			switch typedSlice := c.Input.(type) {
			case []int:
				got = test.AsStrings(typedSlice)
			case []string:
				got = test.AsStrings(typedSlice)
			case []KV:
				got = test.AsStrings(typedSlice)
			default:
				t.Fatalf("unknown slice type: %T", c.Input)
			}
			if !slices.Equal(c.Want, got) {
				t.Errorf("want:\n  %v\ngot:\n  %v", c.Want, got)
			}
		})
	}
}

func TestCheckSlice(t *testing.T) {
	cases := []struct {
		Name      string
		WantInput any
		GotInput  any
		Logs      string
	}{
		{
			Name:      "empties",
			WantInput: []string(nil),
			GotInput:  []string(nil),
		},
		{
			Name:      "same ints",
			WantInput: []int{1, 2, 3},
			GotInput:  []int{1, 2, 3},
		},
		{
			Name:      "same strings",
			WantInput: []string{"hi", "bye"},
			GotInput:  []string{"hi", "bye"},
		},
		{
			Name:      "want one string, get none",
			WantInput: []string{"hi"},
			GotInput:  []string(nil),
			Logs: `slices did not match - missing item at got index 0, skipping rest.
want item:
  hi
last got item:
  <got empty slice>
slices did not match - want(len=1):
  hi
got(len=0):
  <empty slice>
`,
		},
		{
			Name:      "want one string, get different one",
			WantInput: []string{"hi"},
			GotInput:  []string{"bye"},
			Logs: `slices did not match - unexpected item at got index 0:
  bye
want:
  hi
slices did not match - want(len=1):
  hi
got(len=1):
  bye
`,
		},
		{
			Name:      "want two strings, get one",
			WantInput: []string{"hi", "bye"},
			GotInput:  []string{"hi"},
			Logs: `slices did not match - missing item at got index 1, skipping rest.
want item:
  bye
last got item:
  hi
slices did not match - want(len=2):
  hi
  bye
got(len=1):
  hi
`,
		},
		{
			Name:      "want one string, get two",
			WantInput: []string{"hi"},
			GotInput:  []string{"hi", "bye"},
			Logs: `slices did not match - extra item(s) at got index 1, skipping rest.
first extra got item:
  bye
slices did not match - want(len=1):
  hi
got(len=2):
  hi
  bye
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			switch typedWant := c.WantInput.(type) {
			case []string:
				test.CheckSlice(mockT, "slices", typedWant, c.GotInput.([]string))
			case []int:
				test.CheckSlice(mockT, "slices", typedWant, c.GotInput.([]int))
			default:
				t.Fatalf("unknown slice type: %T", c.WantInput)
			}

			var wantFailed bool
			if c.Logs != "" {
				wantFailed = true
			}
			if wantFailed != mockT.Failed() {
				t.Errorf("wantFailed is %v, but t.Failed() is %v",
					wantFailed, mockT.Failed())
			}
			if gotLogs := mockT.buf.String(); c.Logs != gotLogs {
				t.Errorf("logs did not match - want:\n%s\ngot:\n%s", c.Logs, gotLogs)
			}
		})
	}
}
