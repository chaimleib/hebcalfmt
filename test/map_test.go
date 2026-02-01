package test_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
)

func TestCheckMap(t *testing.T) {
	cases := []struct {
		Name string
		Got  map[string]string
		Want map[string]string
		Logs string
	}{
		{Name: "empty"},
		{
			Name: "want empty",
			Got:  map[string]string{"k": "v"},
			Logs: `map did not match - want(len=0):
  <empty map>
got(len=1):
  map[string]string{"k":"v"}
`,
		},
		{
			Name: "got empty",
			Want: map[string]string{"k": "v"},
			Logs: `map did not match - want(len=1):
  map[string]string{"k":"v"}
got(len=0):
  <empty map>
`,
		},
		{
			Name: "gotOnlies",
			Want: map[string]string{"shared": "value"},
			Got:  map[string]string{"shared": "value", "unique": "value"},
			Logs: `map did not match -
	extra values:
		"unique": "value",
`,
		},
		{
			Name: "wantOnlies",
			Want: map[string]string{"shared": "value", "unique": "value"},
			Got:  map[string]string{"shared": "value"},
			Logs: `map did not match -
	missing values:
		"unique": "value",
`,
		},
		{
			Name: "different value",
			Want: map[string]string{"shared": "value A"},
			Got:  map[string]string{"shared": "value B"},
			Logs: `map did not match -
	differing values:
		"shared": {Want: "value A", Got: "value B"},
`,
		},
		{
			Name: "combo diff",
			Want: map[string]string{
				"same":     "value",
				"diff":     "value A",
				"wantOnly": "1",
			},
			Got: map[string]string{
				"same":    "value",
				"diff":    "value B",
				"gotOnly": "0",
			},
			Logs: `map did not match -
	missing values:
		"wantOnly": "1",
	extra values:
		"gotOnly": "0",
	differing values:
		"diff": {Want: "value A", Got: "value B"},
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.CheckMap(mockT, "map", c.Want, c.Got)

			var wantFailed bool
			if c.Logs != "" {
				wantFailed = true
			}
			if wantFailed != mockT.Failed() {
				t.Errorf("wantFailed is %v, but t.Failed() is %v",
					wantFailed, mockT.Failed())
			}
			if gotLogs := mockT.buf.String(); c.Logs != gotLogs {
				t.Errorf("logs did not match - want:\n%q\ngot:\n%q", c.Logs, gotLogs)
			}
		})
	}
}
