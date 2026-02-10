package parsing_test

import (
	"fmt"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
)

func TestContainsByte(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Needle byte
		Want   bool
	}{
		{Name: "empty"},
		{Input: "", Needle: 'a'},
		{Input: "haystack", Needle: 'z'},
		{Input: "haystack", Needle: 'k', Want: true},
		{Input: "hay🤩stack with unicode", Needle: 'c', Want: true},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			var not string
			if !c.Want {
				not = " not"
			}
			name = fmt.Sprintf("%s%s in %s", string(c.Needle), not, c.Input)
		}
		t.Run(c.Name, func(t *testing.T) {
			got := parsing.ContainsByte([]byte(c.Input), c.Needle)
			test.CheckComparable(t, "found", c.Want, got)
		})
	}
}
