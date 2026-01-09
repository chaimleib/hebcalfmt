package test_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
)

func TestCheckString(t *testing.T) {
	cases := []struct {
		Name                string
		WantInput, GotInput string
		Mode                test.WantMode
		Failed              bool
		Logs                string
	}{
		{Name: "empties equal"},
		{Name: "strings equal", WantInput: "hi", GotInput: "hi"},
		{
			Name:      "strings not equal",
			WantInput: "hi",
			GotInput:  "hello",
			Failed:    true,
			Logs: `Field did not match - want:
hi
got:
hello
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
			Logs: `Field did not match prefix - want:
hi
got:
hello
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
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.CheckString(mockT, "Field", c.WantInput, c.GotInput, c.Mode)

			if c.Failed != mockT.Failed() {
				t.Errorf("c.Failed is %v, but t.Failed() is %v",
					c.Failed, mockT.Failed())
			}
			if gotLogs := mockT.buf.String(); c.Logs != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Logs, gotLogs)
			}
		})
	}
}
