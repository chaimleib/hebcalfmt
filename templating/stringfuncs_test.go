package templating_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestTranslate(t *testing.T) {
	cases := []struct {
		Input, Lang, Want string
	}{
		{"Unknown", "en", "Unknown"},
		{"Unknown", "ashkenazi", "Unknown"},
		{"Parashat", "en", "Parashat"},
		{"Parashat", "ashkenazi", "Parshas"},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("To %s: %s", c.Lang, c.Input), func(t *testing.T) {
			got := templating.Translate(c.Lang, c.Input)
			test.CheckString(t, "translation", c.Want, got, test.WantEqual)
		})
	}
}

func TestApply(t *testing.T) {
	cases := []struct {
		Input, Want []string
		Fn          func(string) string
	}{
		{
			Input: []string{"Aleph", "Bet"},
			Want:  []string{"ALEPH", "BET"},
			Fn:    strings.ToUpper,
		},
		{
			Input: []string{"Aleph", "Bet"},
			Want:  []string{"aleph", "bet"},
			Fn:    strings.ToLower,
		},
		{},
	}
	for _, c := range cases {
		t.Run(strings.Join(c.Input, " "), func(t *testing.T) {
			got := templating.Apply(c.Input, c.Fn)
			test.CheckSlice(t, "results", c.Want, got)
		})
	}
}
