package templating_test

import (
	"testing"
	"time"

	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestSecondsDuration(t *testing.T) {
	got := templating.SecondsDuration(300)
	if 5*time.Minute != got {
		t.Errorf("want: 5m, got: %s", got)
	}
}

func TestDatePartsEqual(t *testing.T) {
	cases := []struct {
		Name string
		A, B time.Time
		Want bool
	}{
		{Name: "empties", Want: true},
		{
			Name: "different times of the same day",
			A:    time.Date(1903, 3, 4, 13, 43, 23, 96736, time.UTC),
			B:    time.Date(1903, 3, 4, 3, 23, 43, 322736, time.UTC),
			Want: true,
		},
		{
			Name: "same times",
			A:    time.Date(1903, 3, 4, 13, 43, 23, 96736, time.UTC),
			B:    time.Date(1903, 3, 4, 13, 43, 23, 96736, time.UTC),
			Want: true,
		},
		{
			Name: "different times",
			A:    time.Date(1903, 3, 4, 13, 43, 23, 96736, time.UTC),
			B:    time.Date(1909, 12, 4, 13, 43, 23, 96736, time.UTC),
			Want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := templating.DatePartsEqual(c.A, c.B)
			if c.Want != got {
				t.Errorf("want: %v, got: %v\nA: %s\nB: %s",
					c.Want, got, c.A, c.B)
			}
		})
	}
}

func TestDurationDiv(t *testing.T) {
	cases := []struct {
		Name    string
		D       time.Duration
		Divisor float64
		Want    time.Duration
		Err     string
	}{
		{
			Name:    "Hour by 1",
			D:       time.Hour,
			Divisor: 1,
			Want:    time.Hour,
		},
		{
			Name:    "Zero by 1",
			Divisor: 1,
		},
		{
			Name:    "Hour by 4",
			D:       time.Hour,
			Divisor: 4,
			Want:    15 * time.Minute,
		},
		{
			Name:    "Hour by 1.5",
			D:       time.Hour,
			Divisor: 1.5,
			Want:    40 * time.Minute,
		},
		{
			Name:    "Hour by zero",
			D:       time.Hour,
			Divisor: 0,
			Err:     "divide by zero",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := templating.DurationDiv(c.D, c.Divisor)
			test.CheckErr(t, err, c.Err)
			if got != c.Want {
				t.Errorf("want: %s\ngot:  %s", c.Want, got)
			}
		})
	}
}

func TestDurationMul(t *testing.T) {
	cases := []struct {
		Name   string
		D      time.Duration
		Factor float64
		Want   time.Duration
	}{
		{
			Name:   "Hour by 1",
			D:      time.Hour,
			Factor: 1,
			Want:   time.Hour,
		},
		{
			Name:   "Zero by 1",
			Factor: 1,
		},
		{
			Name:   "Hour by 4",
			D:      time.Hour,
			Factor: 4,
			Want:   4 * time.Hour,
		},
		{
			Name:   "Hour by 1.5",
			D:      time.Hour,
			Factor: 1.5,
			Want:   90 * time.Minute,
		},
		{
			Name:   "Hour by zero",
			D:      time.Hour,
			Factor: 0,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := templating.DurationMul(c.D, c.Factor)
			if got != c.Want {
				t.Errorf("want: %s\ngot:  %s", c.Want, got)
			}
		})
	}
}
