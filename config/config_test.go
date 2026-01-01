package config_test

import (
	"strings"
	"testing"

	"github.com/chaimleib/hebcalfmt/config"
	"github.com/chaimleib/hebcalfmt/daterange"
	"github.com/chaimleib/hebcalfmt/test"
)

func checkConfig(t *testing.T, want, got *config.Config) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("expected nil, got: %#v", got)
		}
		return
	}

fields:
	for _, field := range []struct {
		Name      string
		Want, Got any
	}{
		{"ConfigSource", want.ConfigSource, got.ConfigSource},
		{"DateRange", want.DateRange, got.DateRange},
		{"Now", want.Now, got.Now},
		{"FS", want.FS, got.FS},
		{"Language", want.Language, got.Language},
		{"City", want.City, got.City},
		{"Geo", want.Geo, got.Geo},
		{"Timezone", want.Timezone, got.Timezone},
		// {"Shiurim", want.Shiurim, got.Shiurim}, // TODO: compare string slices
		{"Today", want.Today, got.Today},
		{"ChagOnly", want.ChagOnly, got.ChagOnly},
		{"NoJulian", want.NoJulian, got.NoJulian},
		{"Hour24", want.Hour24, got.Hour24},
		{"SunriseSunset", want.SunriseSunset, got.SunriseSunset},
		{"CandleLighting", want.CandleLighting, got.CandleLighting},
		{"DailyZmanim", want.DailyZmanim, got.DailyZmanim},
		{"Molad", want.Molad, got.Molad},
		{"WeeklyAbbreviated", want.WeeklyAbbreviated, got.WeeklyAbbreviated},
		{"AddHebrewDates", want.AddHebrewDates, got.AddHebrewDates},
		{"AddHebrewDatesForEvents", want.AddHebrewDatesForEvents, got.AddHebrewDatesForEvents},
		{"IsHebrewYear", want.IsHebrewYear, got.IsHebrewYear},
		{"YomKippurKatan", want.YomKippurKatan, got.YomKippurKatan},
		{"ShabbatMevarchim", want.ShabbatMevarchim, got.ShabbatMevarchim},
		{"NoHolidays", want.NoHolidays, got.NoHolidays},
		{"NoRoshChodesh", want.NoRoshChodesh, got.NoRoshChodesh},
		{"IL", want.IL, got.IL},
		{"NoModern", want.NoModern, got.NoModern},
		{"NoMinorFast", want.NoMinorFast, got.NoMinorFast},
		{"NoSpecialShabbat", want.NoSpecialShabbat, got.NoSpecialShabbat},
		{"Omer", want.Omer, got.Omer},
		{"Sedrot", want.Sedrot, got.Sedrot},
		{"DailySedra", want.DailySedra, got.DailySedra},
		{"CandleLightingMins", want.CandleLightingMins, got.CandleLightingMins},
		{"HavdalahMins", want.HavdalahMins, got.HavdalahMins},
		{"HavdalahDeg", want.HavdalahDeg, got.HavdalahDeg},
		{"NumYears", want.NumYears, got.NumYears},
		{"EventsFile", want.EventsFile, got.EventsFile},
		{"YahrzeitsFile", want.YahrzeitsFile, got.YahrzeitsFile},
	} {
		if field.Want != field.Got {
			switch typedWant := field.Want.(type) {
			case *daterange.DateRange:
				if typedWant != nil && field.Got != nil {
					typedGot := field.Got.(*daterange.DateRange)
					test.CheckDateRange(t, *typedWant, *typedGot)
					continue fields
				}

			case *config.Coordinates:
				if typedWant != nil && field.Got != nil {
					typedGot := field.Got.(*config.Coordinates)
					test.CheckCoordinates(t, typedWant, typedGot)
					continue fields
				}

			default:
				t.Errorf("%s's do not match - want:\n%v\ngot:\n%v",
					field.Name, field.Want, field.Got)
			}
		}
	}
}

func TestFromReader(t *testing.T) {
	const fpath = "testConfig.json"

	baseWant := config.Default
	baseWant.ConfigSource = fpath

	cases := []struct {
		Name  string
		Input string
		Want  *config.Config
		Err   string
	}{
		{
			Name:  "invalid empty",
			Input: "",
			Want:  nil,
			Err:   `failed to parse config from "testConfig.json": EOF`,
		},
		{
			Name:  "empty object",
			Input: "{}",
			Want:  &baseWant,
		},
	}
	for _, c := range cases {
		r := strings.NewReader(c.Input)
		got, err := config.FromReader(r, fpath)

		test.CheckErr(t, err, c.Err)
		checkConfig(t, c.Want, got)

	}
}
