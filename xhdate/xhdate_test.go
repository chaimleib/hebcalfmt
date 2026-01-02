package xhdate_test

import (
	"fmt"
	"testing"

	"github.com/hebcal/hdate"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/xhdate"
)

func TestEqual(t *testing.T) {
	var zero hdate.HDate
	rh := hdate.New(5775, hdate.Tishrei, 1)
	rhGreg := rh.Gregorian()
	// rhConverted has cached the RD number,
	// so it is no longer struct-equal with rh
	rhConverted := hdate.FromTime(rhGreg)
	if rh == rhConverted {
		t.Errorf(
			"rh unexpectedly struct-equals rhConverted. Does it still cache the RD?\nrhConverted: %#v",
			rhConverted,
		)
	}
	chanukah := hdate.New(5775, hdate.Kislev, 25)

	cases := []struct {
		Name string
		A, B hdate.HDate
		Want bool
	}{
		{Name: "zeroes", A: zero, B: zero, Want: true},
		{Name: "simple vs zero", A: rh, B: zero, Want: false},
		{Name: "zero vs simple", A: zero, B: rh, Want: false},
		{Name: "simple", A: rh, B: rh, Want: true},
		{Name: "simple vs converted", A: rh, B: rhConverted, Want: true},
		{Name: "converted vs simple", A: rhConverted, B: rh, Want: true},
		{Name: "RH vs Chanukah", A: rh, B: chanukah, Want: false},
		{Name: "Chanukah vs RH", A: chanukah, B: rh, Want: false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := xhdate.Equal(c.A, c.B)
			if c.Want != got {
				t.Errorf("want: %v, got: %v, from %s ?= %s", c.Want, got, c.A, c.B)
			}
		})
	}
}

func TestParse(t *testing.T) {
	rh := hdate.New(5775, hdate.Tishrei, 1)
	purim := hdate.New(5785, hdate.Adar2, 14)
	purimLeap := hdate.New(5787, hdate.Adar2, 14)
	cases := []struct {
		Input string
		Want  hdate.HDate
		Err   string
	}{
		{Input: "Tishrei 1 5775", Want: rh},
		{Input: "1 Tishrei 5775", Want: rh},
		{Input: "tishrei 1 5775", Want: rh},
		{Input: "Ti 1 5775", Want: rh},
		{Input: "1 Ti 5775", Want: rh},
		{Input: "14 Adar 5785", Want: purim},
		{Input: "Adar 14 5785", Want: purim},
		{Input: "14 Adar I 5785", Want: purim},
		{Input: "Adar I 14 5785", Want: purim},
		{Input: "14 Adar II 5785", Want: purim},
		{Input: "Adar II 14 5785", Want: purim},
		{Input: "14 Adar 1 5785", Want: purim},
		{Input: "Adar 1 14 5785", Want: purim},
		{Input: "14 Adar 2 5785", Want: purim},
		{Input: "Adar 2 14 5785", Want: purim},
		{Input: "14 Adar 5787", Want: purimLeap},
		{Input: "Adar 14 5787", Want: purimLeap},
		{Input: "14 Adar2 5787", Want: purimLeap},
		{Input: "Adar2 14 5787", Want: purimLeap},
		{Input: "14 AdarII 5787", Want: purimLeap},
		{Input: "AdarII 14 5787", Want: purimLeap},
		{Input: "14 Adar 2 5787", Want: purimLeap},
		{Input: "Adar 2 14 5787", Want: purimLeap},
		{Input: "14 Adar II 5787", Want: purimLeap},
		{Input: "Adar II 14 5787", Want: purimLeap},

		{Input: "bad date", Err: `too few words in a Hebrew date: "bad date"`},
		{
			Input: "Adar II 14 5787 5788",
			Err:   `too many words in a Hebrew date: "Adar II 14 5787 5788"`,
		},
		{
			Input: "bad month 16 5789",
			Err:   `failed to parse month from Hebrew date: "bad month 16 5789"`,
		},
		{
			Input: "Tishrei 1 badyear",
			Err:   `could not parse last word of Hebrew date as year: "Tishrei 1 badyear"`,
		},
		{
			Input: "Av badday 5234",
			Err:   `could not parse day from Hebrew date: "Av badday 5234"`,
		},
		{
			Input: "5775 Tishrei 1",
			Err:   "invalid day of month for Tishrei 1: 5775",
		},
	}
	for _, c := range cases {
		name := c.Input
		if c.Err != "" {
			name = fmt.Sprintf("invalid - %s", name)
		}
		t.Run(name, func(t *testing.T) {
			got, err := xhdate.Parse(c.Input)
			test.CheckErr(t, err, c.Err)
			if c.Want != got {
				t.Errorf("want: %s, got: %s", c.Want, got)
			}
		})
	}
}
