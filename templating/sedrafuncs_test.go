package templating_test

import (
	"fmt"
	"testing"

	"github.com/hebcal/hdate"

	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestSedra(t *testing.T) {
	const year = 5786
	for _, il := range []bool{false, true} {
		t.Run(fmt.Sprintf("IL=%v", il), func(t *testing.T) {
			sedra := templating.Sedra(year, il)
			p := sedra.Lookup(hdate.New(year, hdate.Adar2, 11))
			if "Parashat Tetzaveh" != p.String() {
				t.Errorf("want: Parashat Tetzaveh, got: %s", p)
			}

			// Do it again to verify that the cache works.
			p2 := sedra.Lookup(hdate.New(year, hdate.Adar2, 11))
			if "Parashat Tetzaveh" != p2.String() {
				t.Errorf("cache - want: Parashat Tetzaveh, got: %s", p2)
			}
		})
	}
}

func TestLocalizedParasha(t *testing.T) {
	cases := []struct {
		HDate hdate.HDate
		IL    bool
		Lang  string
		Want  string
	}{
		{
			HDate: hdate.New(5786, hdate.Tevet, 20),
			Want:  "Parashat Shemot",
		},
		{
			HDate: hdate.New(5786, hdate.Tevet, 20),
			Lang:  "ashkenazi",
			Want:  "Parshas Shemos",
		},
		{
			HDate: hdate.New(5786, hdate.Shvat, 3),
			Want:  "Parashat Bo",
		},
		{
			HDate: hdate.New(5786, hdate.Shvat, 3),
			Lang:  "ashkenazi",
			Want:  "Parshas Bo",
		},
		{
			HDate: hdate.New(5785, hdate.Elul, 25),
			Lang:  "ashkenazi",
			Want:  "Parshas Nitzavim",
		},
		{
			HDate: hdate.New(5784, hdate.Elul, 25),
			Lang:  "ashkenazi",
			Want:  "Parshas Nitzavim-Vayeilech",
		},
		{
			HDate: hdate.New(5784, hdate.Tishrei, 1),
			Lang:  "ashkenazi",
			Want:  "Parshas hachag", // TODO: which chag?
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s (%s)", c.Want, c.Lang), func(t *testing.T) {
			got := templating.LocalizedParasha(c.HDate, c.IL, c.Lang)
			test.CheckString(t, "parasha", c.Want, got)
		})
	}
}
