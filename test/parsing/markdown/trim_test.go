package markdown_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing/markdown"
)

func TestTrimRepeating(t *testing.T) {
	cases := []struct {
		Name  string // defaults to Input
		Input string
		Want  string
	}{
		{Name: "empty"},
		{Input: "hello", Want: "hello"},
		{Input: "llama", Want: "ama"},
		{Input: "```fence", Want: "fence"},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			name = c.Input
		}
		t.Run(name, func(t *testing.T) {
			got, trimLen := markdown.TrimRepeating([]byte(c.Input))
			test.CheckString(t, "trimmed", c.Want, string(got))
			test.CheckComparable(t, "length", len(c.Input)-len(c.Want), trimLen)
		})
	}
}

func TestTrimSpace(t *testing.T) {
	cases := []struct {
		Name  string // defaults to Input
		Input string
		Want  string
	}{
		{Name: "empty"},
		{Input: "hello", Want: "hello"},
		{Input: " hello", Want: "hello"},
		{Input: "\t hello", Want: "hello"},
		{Input: "\n hello", Want: "\n hello"},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			name = c.Input
		}
		t.Run(name, func(t *testing.T) {
			got := markdown.TrimSpace([]byte(c.Input))
			test.CheckString(t, "trimmed", c.Want, string(got))
		})
	}
}
