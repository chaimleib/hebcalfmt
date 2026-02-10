package shell_test

import (
	"strings"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing/shell"
)

func TestTrimSpace(t *testing.T) {
	cases := []struct {
		Name  string // defaults to Input
		Input string
		Want  string
	}{
		{Name: "empty"},
		{Input: "no trim", Want: "no trim"},
		{Input: "  lead trim space", Want: "lead trim space"},
		{Input: "\tlead trim tab", Want: "lead trim tab"},
		{Input: "trail trim space  ", Want: "trail trim space  "},
		{Name: "all spaces", Input: "   "},
		{Name: "all tabs", Input: "\t\t"},
		{Name: "all spaces and tabs", Input: " \t \t"},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			name = strings.TrimSpace(c.Input)
		}
		t.Run(name, func(t *testing.T) {
			got := shell.TrimSpace([]byte(c.Input))
			test.CheckString(t, "output", c.Want, string(got))
		})
	}
}
