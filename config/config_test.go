package config_test

import (
	"errors"
	"io/fs"
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

func TestFromFile(t *testing.T) {
	baseWant := func(fpath string) *config.Config {
		cfg := config.Default
		cfg.ConfigSource = fpath
		return &cfg
	}

	cases := []struct {
		Name  string
		FSErr error
		Fpath string
		Want  func(fpath string) *config.Config
		Err   string
	}{
		{
			Name:  "invalid empty file",
			Fpath: "testdata/empty.txt",
			Want:  nil,
			Err:   `failed to parse config from "testdata/empty.txt": EOF`,
		},
		{
			Name:  "invalid nonexistent file",
			Fpath: "testdata/nonexistent.json",
			Want:  nil,
			Err:   `config file could not be read: open testdata/nonexistent.json: no such file or directory`,
		},
		{
			Name:  "invalid FS",
			FSErr: errors.New("test forced DefaultFS to fail"),
			Fpath: "testdata/emptyObject.json",
			Want:  nil,
			Err:   `failed to initialize DefaultFS: test forced DefaultFS to fail`,
		},
		{
			Name:  "empty object",
			Fpath: "testdata/emptyObject.json",
			Want:  baseWant,
		},
		{
			Name:  "today config",
			Fpath: "testdata/today.json",
			Want: func(fpath string) *config.Config {
				cfg := baseWant(fpath)
				cfg.Today = true
				return cfg
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.FSErr != nil {
				old := config.DefaultFS
				t.Cleanup(func() {
					config.DefaultFS = old
				})
				config.DefaultFS = func() (fs.FS, error) {
					return nil, c.FSErr
				}
			}

			got, err := config.FromFile(c.Fpath)
			var want *config.Config
			if c.Want != nil {
				want = c.Want(c.Fpath)
			}

			test.CheckErr(t, err, c.Err)
			checkConfig(t, want, got)
		})
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
		t.Run(c.Name, func(t *testing.T) {
			r := strings.NewReader(c.Input)
			got, err := config.FromReader(r, fpath)

			test.CheckErr(t, err, c.Err)
			checkConfig(t, c.Want, got)
		})
	}
}

func TestNormalize(t *testing.T) {
	baseWant := func() *config.Config {
		cfg := config.Default
		return &cfg
	}

	cases := []struct {
		Name string
		Cfg  *config.Config
		Want *config.Config
		Err  string
		Log  string
	}{
		{
			Name: "empty",
			Cfg:  new(config.Config),
			Want: &config.Config{Language: "en"},
		},
		{
			Name: "base config",
			Cfg:  baseWant(),
			Want: func() *config.Config {
				cfg := baseWant()
				cfg.Language = "en"
				return cfg
			}(),
		},
		{
			Name: "language en",
			Cfg:  &config.Config{Language: "en"},
			Want: &config.Config{Language: "en"},
		},
		{
			Name: "language ASHKENAZI",
			Cfg:  &config.Config{Language: "ASHKENAZI"},
			Want: &config.Config{Language: "ashkenazi"},
		},
		{
			Name: "language invalid",
			Cfg:  &config.Config{Language: "invalid"},
			Want: nil,
			Err:  `unknown language: "invalid"`,
			Log: strings.TrimSpace(`
unknown language: "invalid"
To show the available languages, run
  hebcalfmt --info languages
			`),
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			buf := test.TestLogger(t)
			got, err := c.Cfg.Normalize()
			test.CheckErr(t, err, c.Err)
			checkConfig(t, c.Want, got)
			if c.Log != strings.TrimSpace(buf.String()) {
				t.Errorf("want logs:\n%s\ngot logs:\n%s", c.Log, buf)
			}
		})
	}
}
