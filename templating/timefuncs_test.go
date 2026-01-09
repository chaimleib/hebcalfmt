package templating_test

import (
	"testing"
	"time"

	"github.com/chaimleib/hebcalfmt/templating"
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
