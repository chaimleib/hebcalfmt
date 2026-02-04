package test_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
)

func TestCheckString(t *testing.T) {
	type Case struct {
		Name                string
		WantInput, GotInput string
		Mode                test.WantMode
		Failed              bool
		Logs                string
	}
	cases := []Case{
		{Name: "empties equal"},
		{Name: "strings equal", WantInput: "hi", GotInput: "hi"},
		{
			Name:      "multiline strings equal",
			WantInput: "hi\nbye",
			GotInput:  "hi\nbye",
		},
		{
			Name:      "strings not equal",
			WantInput: "hi",
			GotInput:  "hello",
			Failed:    true,
			Logs: `Field did not match at index 1 (line 1, col 2) -
want: hi
got:  hello
       ^
`,
		},
		{
			Name:      "strings not equal with longer want",
			WantInput: "hello there",
			GotInput:  "hello",
			Failed:    true,
			Logs: `Field did not match at index 5 (line 1, col 6) -
want: hello there
got:  hello
           ^
`,
		},
		{
			Name:      "strings not equal with longer got",
			WantInput: "hello",
			GotInput:  "hello there",
			Failed:    true,
			Logs: `Field did not match at index 5 (line 1, col 6) -
want: hello
got:  hello there
           ^
`,
		},
		{
			Name:      "strings not equal with got trailing newline",
			WantInput: "hi",
			GotInput:  "hi\n",
			Failed:    true,
			Logs: `Field did not match at index 2 (line 1, col 3) -
want: hi
got:  hi⏎
        ^
`,
		},
		{
			Name:      "strings not equal with want trailing newline",
			WantInput: "hi\n",
			GotInput:  "hi",
			Failed:    true,
			Logs: `Field did not match at index 2 (line 1, col 3) -
want: hi⏎
got:  hi
        ^
`,
		},
		{
			Name:      "multiline strings not equal",
			WantInput: "hello\nbye",
			GotInput:  "hello\nbye-bye",
			Failed:    true,
			Logs: `Field did not match at index 9 (line 2, col 4) -
want: bye
got:  bye-bye
         ^
`,
		},
		{
			Name:      "tabbed strings not equal",
			WantInput: "hello\tbye",
			GotInput:  "hello\tbye-bye",
			Failed:    true,
			Logs: `Field did not match at index 9 (line 1, col 10) -
want: hello   bye
got:  hello   bye-bye
                 ^
`,
		},

		{
			Name:      "strings prefix ok",
			WantInput: "hi",
			GotInput:  "high",
			Mode:      test.WantPrefix,
		},
		{
			Name:      "strings prefix fail",
			WantInput: "hi",
			GotInput:  "hello",
			Mode:      test.WantPrefix,
			Failed:    true,
			Logs: `Field did not match prefix
Field did not match at index 1 (line 1, col 2) -
want: hi
got:  hello
       ^
`,
		},

		{
			Name:      "strings contains ok",
			WantInput: "ig",
			GotInput:  "high",
			Mode:      test.WantContains,
		},
		{
			Name:      "strings contains fail",
			WantInput: "NOPE",
			GotInput:  "hello",
			Mode:      test.WantContains,
			Failed:    true,
			Logs: `Field did not contain string - want:
NOPE
got:
hello
`,
		},

		{
			Name:      "strings regex ok",
			WantInput: `([a-z])igh(?:light)?`,
			GotInput:  "high",
			Mode:      test.WantRegexp,
		},
		{
			Name:      "strings regex fail",
			WantInput: ".{6}",
			GotInput:  "hello",
			Mode:      test.WantRegexp,
			Failed:    true,
			Logs: `Field did not match regexp - want:
.{6}
got:
hello
`,
		},

		{
			Name:      "strings ellipsis ok with len 1 splits",
			WantInput: `high`,
			GotInput:  "high",
			Mode:      test.WantEllipsis,
		},
		{
			Name:      "strings ellipsis fail with len 1 splits",
			WantInput: "bye",
			GotInput:  "hello",
			Mode:      test.WantEllipsis,
			Failed:    true,
			Logs: `Field did not match ellipsis portion 1 of 1 -
Field did not match at index 0 (line 1, col 1) -
want: bye
got:  hello
      ^
`,
		},
		{
			Name:      "strings ellipsis fail trailing got with len 1 splits",
			WantInput: "hello",
			GotInput:  "hello there",
			Mode:      test.WantEllipsis,
			Failed:    true,
			Logs: `Field did not match, has trailing content after wanted string -
Field did not match at index 5 (line 1, col 6) -
want: hello
got:  hello there
           ^
`,
		},
		{
			Name:      "strings ellipsis ok with len 2 splits",
			WantInput: `high...five`,
			GotInput:  "high friggin five",
			Mode:      test.WantEllipsis,
		},
		{
			Name:      "strings ellipsis fail with len 2 splits on split 0",
			WantInput: "bye... later",
			GotInput:  "hello",
			Mode:      test.WantEllipsis,
			Failed:    true,
			Logs: `Field did not match ellipsis portion 1 of 2 -
Field did not match at index 0 (line 1, col 1) -
want: bye
got:  hello
      ^
`,
		},
		{
			Name:      "strings ellipsis fail with len 2 splits on split 1",
			WantInput: "bye... later",
			GotInput:  "bye, see you some other time",
			Mode:      test.WantEllipsis,
			Failed:    true,
			Logs: `Field did not match, ellipsis portion 2 of 2 not found -
want:
 later
somewhere at or after got[3:] (line 1, col 4):
bye, see you some other time
   ^
`,
		},
		{
			Name:      "strings ellipsis fail trailing got with len 2 splits",
			WantInput: "hello... there",
			GotInput:  "hello over there, little one",
			Mode:      test.WantEllipsis,
			Failed:    true,
			Logs: `Field did not match, unexpected trailing content after last ellipsis portion -
got[16:] (line 1, col 17):
hello over there, little one
                ^
`,
		},
		{
			Name:      "strings ellipsis fail trailing multiline got with len 2 splits",
			WantInput: "hello... there",
			GotInput:  "hello over there, little one\nthe world is large",
			Mode:      test.WantEllipsis,
			Failed:    true,
			Logs: `Field did not match, unexpected trailing content after last ellipsis portion -
got[16:] (line 1, col 17):
hello over there, little one⏎...
                ^
`,
		},
		{
			Name: "strings ellipsis fail with len 3 splits on split 1",
			WantInput: `alpha
...
charlie
...
echo`,
			GotInput: `alpha
bravo
c
delta
echo`,
			Mode:   test.WantEllipsis,
			Failed: true,
			Logs: `Field did not match, ellipsis portion 2 of 3 not found -
want:

charlie

somewhere at or after got[6:] (line 2, col 1):
bravo⏎...
^
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.CheckStringMode(mockT, "Field", c.WantInput, c.GotInput, c.Mode)

			if c.Failed != mockT.Failed() {
				t.Errorf("c.Failed is %v, but t.Failed() is %v",
					c.Failed, mockT.Failed())
			}
			gotLogs := mockT.buf.String()
			if c.Logs != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Logs, gotLogs)
			}
		})
	}
}

func TestShowFirstDiff(t *testing.T) {
	type Case struct {
		Name   string
		Want   string
		Got    string
		Offset int
		Logs   string
	}
	cases := []Case{
		{Name: "empty"},
		{
			Name:   "negative offset",
			Offset: -1,
			Logs: `ShowFirstDiff offset out of bounds: -1
`,
		},
		{
			Name:   "offset exceeds string len",
			Offset: 1000,
			Got:    "hello",
			Logs: `ShowFirstDiff offset out of bounds: 1000
`,
		},
		{
			Name:   "ok offset equals string len",
			Offset: 5,
			Got:    "hello",
		},
		{
			Name:   "fail offset equals string len",
			Offset: 5,
			Got:    "hello",
			Want:   "not at offset 5",
			Logs: `Field did not match at index 5 (line 1, col 6) -
     want: not at offset 5
got:  hello
           ^
`,
		},
		{
			Name:   "fail offset equals string len after newline",
			Offset: 6,
			Got:    "hello\n",
			Want:   "not at offset 6",
			Logs: `Field did not match at index 6 (line 1, col 7) -
      want: not at offset 6
got:  hello⏎
            ^
`,
		},
		{
			Name:   "offset strings not equal with got trailing newline",
			Offset: 4,
			Want:   "hi",
			Got:    "oh, hi\n",
			Logs: `Field did not match at index 6 (line 1, col 7) -
    want: hi
got:  oh, hi⏎
            ^
`,
		},
		{
			Name:   "offset strings not equal with want trailing newline",
			Offset: 4,
			Want:   "hi\n",
			Got:    "oh, hi",
			Logs: `Field did not match at index 6 (line 1, col 7) -
    want: hi⏎
got:  oh, hi
            ^
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.ShowFirstDiff(mockT, "Field", c.Want, c.Got, c.Offset)

			wantFailed := c.Logs != ""
			if wantFailed != mockT.Failed() {
				t.Errorf("wantFailed is %v, but t.Failed() is %v",
					wantFailed, mockT.Failed())
			}
			if gotLogs := mockT.buf.String(); c.Logs != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Logs, gotLogs)
			}
		})
	}
}
