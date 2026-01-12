package templating_test

import (
	"testing"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/hebcal"

	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestSetDates(t *testing.T) {
	cases := []struct {
		Name  string
		Dates []hdate.HDate
		Opts  hebcal.CalOptions
		Want  hebcal.CalOptions
		Err   string
	}{
		{Name: "empties"},
		{
			Name:  "one date",
			Dates: []hdate.HDate{hdate.New(5630, hdate.Nisan, 2)},
			Want: hebcal.CalOptions{
				NumYears: 1,
				Start:    hdate.New(5630, hdate.Nisan, 2),
				End:      hdate.New(5630, hdate.Nisan, 2),
			},
		},
		{
			Name: "two dates",
			Dates: []hdate.HDate{
				hdate.New(5630, hdate.Nisan, 2),
				hdate.New(5630, hdate.Nisan, 9),
			},
			Want: hebcal.CalOptions{
				NumYears: 1,
				Start:    hdate.New(5630, hdate.Nisan, 2),
				End:      hdate.New(5630, hdate.Nisan, 9),
			},
		},
		{
			Name: "out of order",
			Dates: []hdate.HDate{
				hdate.New(5630, hdate.Nisan, 9),
				hdate.New(5630, hdate.Nisan, 2),
			},
			Err: "first date must be before second, got 9 Nisan 5630/1870-04-10, 2 Nisan 5630/1870-04-03",
		},
		{
			Name: "too many dates",
			Dates: []hdate.HDate{
				hdate.New(5630, hdate.Nisan, 2),
				hdate.New(5630, hdate.Nisan, 9),
				hdate.New(5630, hdate.Nisan, 9),
			},
			Err: "expected 0-2 dates, got 3",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			gotOpts := c.Opts
			got, err := templating.SetDates(&gotOpts)(c.Dates...)
			test.CheckErr(t, err, c.Err)
			test.CheckString(
				t, "dummy return", "", got.(string), test.WantEqual)
		})
	}
}

func TestSetStart(t *testing.T) {
	got := hebcal.CalOptions{
		NumYears: 99,
		Year:     -1,
	}
	dummy := templating.SetStart(&got)(hdate.New(5720, hdate.Sivan, 6))
	if dummy.(string) != "" {
		t.Errorf("unexpected value was returned: %#v", dummy)
	}
	want := hebcal.CalOptions{
		NumYears: 1,
		Year:     0,
		Start:    hdate.New(5720, hdate.Sivan, 6),
	}
	test.CheckCalOptions(t, &want, &got)
}

func TestSetEnd(t *testing.T) {
	got := hebcal.CalOptions{
		NumYears: 99,
		Year:     -1,
	}
	dummy := templating.SetEnd(&got)(hdate.New(5720, hdate.Sivan, 6))
	if dummy.(string) != "" {
		t.Errorf("unexpected value was returned: %#v", dummy)
	}
	want := hebcal.CalOptions{
		NumYears: 1,
		Year:     0,
		End:      hdate.New(5720, hdate.Sivan, 6),
	}
	test.CheckCalOptions(t, &want, &got)
}

func TestSetYear(t *testing.T) {
	got := hebcal.CalOptions{
		Year:  -1,
		Start: hdate.New(5720, hdate.Sivan, 6),
		End:   hdate.New(5720, hdate.Sivan, 6),
	}
	dummy := templating.SetYear(&got)(1976)
	if dummy.(string) != "" {
		t.Errorf("unexpected value was returned: %#v", dummy)
	}
	want := hebcal.CalOptions{
		Year: 1976,
	}
	test.CheckCalOptions(t, &want, &got)
}

func TestSetNumYears(t *testing.T) {
	got := hebcal.CalOptions{
		NumYears: 10,
	}
	dummy := templating.SetNumYears(&got)(5)
	if dummy.(string) != "" {
		t.Errorf("unexpected value was returned: %#v", dummy)
	}
	want := hebcal.CalOptions{
		NumYears: 5,
	}
	test.CheckCalOptions(t, &want, &got)
}

func TestSetIsHebrewYear(t *testing.T) {
	got := hebcal.CalOptions{
		IsHebrewYear: false,
	}
	dummy := templating.SetIsHebrewYear(&got)(true)
	if dummy.(string) != "" {
		t.Errorf("unexpected value was returned: %#v", dummy)
	}
	want := hebcal.CalOptions{
		IsHebrewYear: true,
	}
	test.CheckCalOptions(t, &want, &got)
}
