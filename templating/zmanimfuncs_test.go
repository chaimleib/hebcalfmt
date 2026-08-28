package templating_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestForLocationDate(t *testing.T) {
	chicago := &zmanim.Location{TimeZoneId: "America/Chicago"}
	chicagoTZ, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Error(err)
	}

	cases := []struct {
		Name         string
		Date         time.Time
		Location     *zmanim.Location
		UseElevation bool
		Want         *zmanim.Zmanim
		Err          string
	}{
		{Name: "empty", Err: "provided location was nil"},
		{
			Name:         "ok",
			Date:         time.Date(1980, time.May, 3, 0, 0, 0, 0, time.UTC),
			Location:     chicago,
			UseElevation: false,
			Want: &zmanim.Zmanim{
				Location:     chicago,
				Year:         1980,
				Month:        time.May,
				Day:          3,
				TimeZone:     chicagoTZ,
				UseElevation: false,
			},
		},
		{
			Name: "invalid TZ", Location: &zmanim.Location{
				TimeZoneId: "INVALID ZONE",
			},
			Err: "unknown time zone INVALID ZONE",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := templating.ForLocationDate(c.Location, c.UseElevation, c.Date)
			test.CheckErr(t, err, c.Err)

			if c.Want == nil && got == nil {
				return
			} else if c.Want == nil {
				t.Errorf("want nil, got: %#v", got)
				return
			} else if got == nil {
				t.Errorf("got nil, want: %#v", c.Want)
				return
			}

			// Check Zmanim equality
			if !reflect.DeepEqual(c.Want.Location, got.Location) {
				t.Errorf(
					"Locations do not match - want:\n  %#v\ngot:\n  %#v",
					c.Want.Location,
					got.Location,
				)
			}
			test.CheckComparable(t, "Year", c.Want.Year, got.Year)
			test.CheckComparable(t, "Month", c.Want.Month, got.Month)
			test.CheckComparable(t, "Day", c.Want.Day, got.Day)
			if !reflect.DeepEqual(c.Want.TimeZone, got.TimeZone) {
				t.Errorf(
					"TimeZones do not match - want:\n  %#v\ngot:\n  %#v",
					c.Want.TimeZone,
					got.TimeZone,
				)
			}
			test.CheckComparable(
				t,
				"UseElevation",
				c.Want.UseElevation,
				got.UseElevation,
			)
		})
	}
}

func TestLookupCity(t *testing.T) {
	cases := []struct {
		Name string
		Err  string
	}{
		{Name: "New York"},
		{Name: "Austin"},
		{Name: "Invalid City", Err: `unknown city "Invalid City"`},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := templating.LookupCity(c.Name)
			test.CheckErr(t, err, c.Err)
			if c.Err == "" && got == nil {
				t.Error("no error expected, but got no city!")
			}
		})
	}
}
