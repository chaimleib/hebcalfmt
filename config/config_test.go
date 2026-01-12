package config_test

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/yerushalmi"
	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/config"
	"github.com/chaimleib/hebcalfmt/daterange"
	"github.com/chaimleib/hebcalfmt/fsys"
	"github.com/chaimleib/hebcalfmt/test"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func checkConfig(t *testing.T, want, got *config.Config) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("expected nil, got: %#v", got)
		}
		return
	}
	if got == nil {
		t.Errorf("got nil, want: %#v", want)
		return
	}

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
		{"Shiurim", want.Shiurim, got.Shiurim},
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
		// In case the type is an interface (doesn't work for pointers),
		// check for mismatching nilness or matching nils.
		// If we proceed past both checks,
		// both are non-nil and castable to the interface type.
		if (field.Want == nil) != (field.Got == nil) {
			t.Errorf("%s's did not match - want:\n%#v\ngot:\n%#v",
				field.Name, field.Want, field.Got)
			continue
		} else if field.Want == nil { // implies field.Got == nil b/c of prev check
			continue
		}

		switch typedWant := field.Want.(type) {
		case fs.FS:
			// assumes field.Got != nil
			typedGot, ok := field.Got.(fs.FS)
			if !ok {
				t.Errorf("%s's types did not match - want:\n%T\ngot:\n%T",
					field.Name, typedWant, field.Got)
			} else {
				test.CheckFS(t, field.Name, typedWant, typedGot)
			}

		case *daterange.DateRange:
			test.CheckNilPtrThen(t, test.CheckDateRange, field.Name, typedWant, field.Got)

		case *config.Coordinates:
			test.CheckNilPtrThen(t, test.CheckCoordinates, field.Name, typedWant, field.Got)

		case []string:
			typedGot := field.Got.([]string)
			if !slices.Equal(typedWant, typedGot) {
				t.Errorf("%s's do not match - want:\n%v\ngot:\n%v",
					field.Name, field.Want, field.Got)
			}

		default:
			test.CheckComparable(t, field.Name, field.Want, field.Got)
		}
	}
}

func TestFromFile(t *testing.T) {
	files := fstest.MapFS{
		"empty.txt":        &fstest.MapFile{Data: []byte("")},
		"emptyObject.json": &fstest.MapFile{Data: []byte("{}")},
		"today.json":       &fstest.MapFile{Data: []byte(`{"today":true}`)},
	}
	baseWant := func(fpath string) *config.Config {
		cfg := config.Default
		cfg.ConfigSource = fpath
		cfg.FS = fsys.WrapFS{
			BaseDir: filepath.Dir(fpath),
			FS:      files,
		}
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
			Fpath: "empty.txt",
			Want:  nil,
			Err:   `failed to parse config from "empty.txt": EOF`,
		},
		{
			Name:  "invalid nonexistent file",
			Fpath: "nonexistent.json",
			Want:  nil,
			Err:   `config file could not be read: open nonexistent.json: file does not exist`,
		},
		{
			Name:  "custom FS",
			FSErr: errors.New("test forced files.Open to fail"),
			Fpath: "emptyObject.json",
			Want:  nil,
			Err:   `config file could not be read: test forced files.Open to fail`,
		},
		{
			Name:  "empty object",
			Fpath: "emptyObject.json",
			Want:  baseWant,
		},
		{
			Name:  "today config",
			Fpath: "today.json",
			Want: func(fpath string) *config.Config {
				cfg := baseWant(fpath)
				cfg.Today = true
				return cfg
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			test.Slogger(t)
			test.Logger(t)

			var files fs.FS = files
			if c.FSErr != nil {
				openReturnsErr := func(string) (fs.File, error) {
					return nil, c.FSErr
				}
				files = fsys.NewFSFunc(openReturnsErr)
			}

			got, err := config.FromFile(files, c.Fpath)
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
			buf := test.Logger(t)
			got, err := c.Cfg.Normalize()
			test.CheckErr(t, err, c.Err)
			checkConfig(t, c.Want, got)
			if c.Log != strings.TrimSpace(buf.String()) {
				t.Errorf("want logs:\n%s\ngot logs:\n%s", c.Log, buf)
			}
		})
	}
}

func TestCalOptions(t *testing.T) {
	// baseWant := func() *hebcal.CalOptions {
	// 	opts := hebcal.CalOptions{
	// 		CandleLightingMins: 18,
	// 		NumYears: 1,
	// 	}
	// 	return &opts
	// }
	nyc := &zmanim.Location{
		Name:        "New York",
		CountryCode: "US",
		Latitude:    40.71427,
		Longitude:   -74.00597,
		TimeZoneId:  "America/New_York",
	}
	failingFS := func() (fs.FS, error) {
		return nil, errors.New("test forced failure to init FS")
	}
	file := func(s string) *fstest.MapFile {
		return &fstest.MapFile{Data: []byte(s)}
	}
	files := fstest.MapFS{
		"events.txt":    file("Tishrei 2 Birthday - Ben Ploni"),
		"yahrzeits.txt": file("10 8 1967 Yahrzeit - Joe Shmo"),
	}

	cases := []struct {
		Name   string
		Cfg    config.Config
		Want   *hebcal.CalOptions
		FailFS bool
		Err    string
	}{
		{Name: "empty", Want: &hebcal.CalOptions{Year: 1, Location: nyc}},
		{
			Name: "explicit city",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.City = "New York"
				return cfg
			}(),
			Want: &hebcal.CalOptions{
				Year:               1,
				NumYears:           1,
				CandleLightingMins: 18,
				HavdalahMins:       72,
				CandleLighting:     true,
				Location:           nyc,
			},
		},
		{
			Name: "today",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.Today = true
				return cfg
			}(),
			Want: &hebcal.CalOptions{
				Year:               1,
				NumYears:           1,
				CandleLightingMins: 18,
				Location:           nyc,
				Omer:               true,
				AddHebrewDates:     true,
			},
		},
		{
			Name: "chag-only",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.ChagOnly = true
				return cfg
			}(),
			Want: &hebcal.CalOptions{
				Year:               1,
				NumYears:           1,
				CandleLightingMins: 18,
				Location:           nyc,
				Mask: event.CHAG | event.LIGHT_CANDLES |
					event.LIGHT_CANDLES_TZEIS | event.YOM_TOV_ENDS,
			},
		},
		{
			Name: "yahrzeit file",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.YahrzeitsFile = "yahrzeits.txt"
				cfg.FS = files
				return cfg
			}(),
			Want: &hebcal.CalOptions{
				Year:               1,
				NumYears:           1,
				CandleLightingMins: 18,
				Location:           nyc,
				Yahrzeits: []hebcal.UserYahrzeit{
					{
						Name: "Yahrzeit - Joe Shmo",
						Date: date(1967, time.October, 8),
					},
				},
			},
		},
		{
			Name: "event file",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.EventsFile = "events.txt"
				cfg.FS = files
				return cfg
			}(),
			Want: &hebcal.CalOptions{
				Year:               1,
				NumYears:           1,
				CandleLightingMins: 18,
				Location:           nyc,
				UserEvents: []hebcal.UserEvent{
					{
						Desc:  "Birthday - Ben Ploni",
						Day:   2,
						Month: hdate.Tishrei,
					},
				},
			},
		},

		// Errors
		{
			Name: "unknown city",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.City = "Unknown"
				return cfg
			}(),
			Err: `failed to resolve place configs: unknown city: "Unknown"`,
		},
		{
			Name: "invalid DateRange",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.DateRange = new(daterange.DateRange)
				return cfg
			}(),
			Err: "range type is YEAR, but the date provided is missing the year: DateRange<empty>",
		},
		{
			Name: "unknown Shiurim",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.Shiurim = []string{"unknown"}
				return cfg
			}(),
			Err: `unrecognized item(s) in shiurim: ["unknown"]`,
		},
		{
			Name:   "fail fs",
			FailFS: true,
			Err:    "failed to initialize DefaultFS: test forced failure to init FS",
		},
		{
			Name: "invalid yahrzeit file",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.YahrzeitsFile = "nonexistent.txt"
				cfg.FS = files
				return cfg
			}(),
			Err: "open nonexistent.txt: file does not exist",
		},
		{
			Name: "invalid event file",
			Cfg: func() config.Config {
				cfg := config.Default
				cfg.EventsFile = "nonexistent.txt"
				cfg.FS = files
				return cfg
			}(),
			Err: "open nonexistent.txt: file does not exist",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			test.Logger(t)
			if c.FailFS {
				old := fsys.DefaultFS
				t.Cleanup(func() { fsys.DefaultFS = old })
				fsys.DefaultFS = failingFS
			}
			got, err := c.Cfg.CalOptions()
			test.CheckErr(t, err, c.Err)
			if c.Err == "" { // otherwise, don't care
				test.CheckCalOptions(t, c.Want, got)
			}
		})
	}
}

func TestSetDateRange(t *testing.T) {
	now := date(2026, 5, 2)
	cases := []struct {
		Name string
		Cfg  *config.Config
		Want *hebcal.CalOptions
		Err  string
	}{
		{Name: "empty", Cfg: new(config.Config), Want: &hebcal.CalOptions{Year: 1}},

		{
			Name: "DateRange of today",
			Cfg: &config.Config{
				Today:    true,
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source:    daterange.Source{Now: now},
					RangeType: daterange.RangeTypeToday,
					Year:      2026,
					GregMonth: time.May,
					Day:       2,
				},
			},
			Want: &hebcal.CalOptions{
				Start:          hdate.New(5786, hdate.Iyyar, 15),
				End:            hdate.New(5786, hdate.Iyyar, 15),
				AddHebrewDates: true,
			},
		},

		{
			Name: "DateRange of Gregorian year",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source:    daterange.Source{Now: now},
					RangeType: daterange.RangeTypeYear,
					Year:      2026,
				},
			},
			Want: &hebcal.CalOptions{Year: 2026},
		},
		{
			Name: "DateRange of Gregorian month",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source:    daterange.Source{Args: []string{"2", "2030"}, Now: now},
					RangeType: daterange.RangeTypeMonth,
					Year:      2030,
					GregMonth: time.February,
				},
			},
			Want: &hebcal.CalOptions{
				Start: hdate.New(5790, hdate.Shvat, 28),
				End:   hdate.New(5790, hdate.Adar1, 25),
			},
		},
		{
			Name: "DateRange of Gregorian day",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source: daterange.Source{
						Args: []string{"2", "29", "2032"},
						Now:  now,
					},
					RangeType: daterange.RangeTypeDay,
					Year:      2032,
					GregMonth: time.February,
					Day:       29,
				},
			},
			Want: &hebcal.CalOptions{
				Start:          hdate.New(5792, hdate.Adar1, 17),
				End:            hdate.New(5792, hdate.Adar1, 17),
				AddHebrewDates: true,
			},
		},

		{
			Name: "DateRange of Hebrew year",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source:       daterange.Source{Now: now},
					RangeType:    daterange.RangeTypeYear,
					Year:         5770,
					IsHebrewDate: true,
				},
			},
			Want: &hebcal.CalOptions{Year: 5770, IsHebrewYear: true},
		},
		{
			Name: "DateRange of Hebrew month",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source: daterange.Source{
						Args: []string{"Elul", "5771"},
						Now:  now,
					},
					RangeType:    daterange.RangeTypeMonth,
					Year:         5771,
					HebMonth:     hdate.Elul,
					IsHebrewDate: true,
				},
			},
			Want: &hebcal.CalOptions{
				Start:        hdate.New(5771, hdate.Elul, 1),
				End:          hdate.New(5771, hdate.Elul, 29),
				IsHebrewYear: true,
			},
		},
		{
			Name: "DateRange of Hebrew day",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source: daterange.Source{
						Args: []string{"Iyar", "1", "5772"},
						Now:  now,
					},
					RangeType:    daterange.RangeTypeDay,
					Year:         5772,
					HebMonth:     hdate.Iyyar,
					Day:          1,
					IsHebrewDate: true,
				},
			},
			Want: &hebcal.CalOptions{
				Start:          hdate.New(5772, hdate.Iyyar, 1),
				End:            hdate.New(5772, hdate.Iyyar, 1),
				IsHebrewYear:   true,
				AddHebrewDates: true,
			},
		},

		{
			Name: "DateRange of year with invalid NumYears",
			Cfg: &config.Config{
				NumYears: 0,
				DateRange: &daterange.DateRange{
					Source:    daterange.Source{Now: now},
					RangeType: daterange.RangeTypeYear,
					Year:      2026,
				},
			},
			Err: "invalid num_years: 0",
		},
		{
			Name: "DateRange of month with too many NumYears",
			Cfg: &config.Config{
				NumYears: 2,
				DateRange: &daterange.DateRange{
					Source:    daterange.Source{Args: []string{"2", "2030"}, Now: now},
					RangeType: daterange.RangeTypeMonth,
					Year:      2030,
					GregMonth: time.February,
				},
			},
			Err: "num_years was 2, but the parsed date range spec was DateRange<February 2030>, not just a year",
		},

		{
			Name: "DateRange of month with Today flag",
			Cfg: &config.Config{
				Today:    true,
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source: daterange.Source{
						Args: []string{"5", "2026"},
						Now:  now,
					},
					RangeType: daterange.RangeTypeMonth,
					Year:      2026,
					GregMonth: time.May,
				},
			},
			Err: "today option works only with single-day calendars, but date range spec was DateRange<May 2026>",
		},

		{
			Name: "DateRange of day with missing day",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source: daterange.Source{
						Args: []string{"2", "2032"},
						Now:  now,
					},
					RangeType: daterange.RangeTypeDay,
					Year:      2032,
					GregMonth: time.February,
				},
			},
			Err: `range type is DAY, but the date provided is missing the day of the month: DateRange<0 February 2032>`,
		},
		{
			Name: "DateRange of month with missing month",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source:    daterange.Source{Args: []string{"2030"}, Now: now},
					RangeType: daterange.RangeTypeMonth,
					Year:      2030,
				},
			},
			Err: `range type is MONTH, but the Gregorian date provided is missing the month: DateRange<%!Month(0) 2030>`,
		},
		{
			Name: "DateRange of year with missing year",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source:    daterange.Source{Now: now},
					RangeType: daterange.RangeTypeYear,
				},
			},
			Err: "range type is YEAR, but the date provided is missing the year: DateRange<0>",
		},

		{
			Name: "DateRange of Hebrew year with missing year",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source:       daterange.Source{Now: now},
					RangeType:    daterange.RangeTypeYear,
					IsHebrewDate: true,
				},
			},
			Err: "range type is YEAR, but the date provided is missing the year: DateRange<0 (Hebrew)>",
		},
		{
			Name: "DateRange of Hebrew month with missing month",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source: daterange.Source{
						Args: []string{"5771"},
						Now:  now,
					},
					RangeType:    daterange.RangeTypeMonth,
					Year:         5771,
					IsHebrewDate: true,
				},
			},
			Err: "range type is MONTH, but the Hebrew date provided is missing the month: DateRange<%!HMonth(0) 5771>",
		},
		{
			Name: "DateRange of Hebrew day with missing day",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source: daterange.Source{
						Args: []string{"Iyar", "5772"},
						Now:  now,
					},
					RangeType:    daterange.RangeTypeDay,
					Year:         5772,
					HebMonth:     hdate.Iyyar,
					IsHebrewDate: true,
				},
			},
			Err: "range type is DAY, but the date provided is missing the day of the month: DateRange<0 Iyyar 5772>",
		},

		{
			Name: "DateRange of unknown RangeType",
			Cfg: &config.Config{
				NumYears: 1,
				DateRange: &daterange.DateRange{
					Source:    daterange.Source{Now: now},
					RangeType: -1,
				},
			},
			Err: "unreachable code: invalid RangeType value: UNKNOWN(-1)",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			want := c.Want
			var got hebcal.CalOptions
			err := c.Cfg.SetDateRange(&got)
			test.CheckErr(t, err, c.Err)
			if c.Err == "" { // otherwise if err, don't care about got
				test.CheckCalOptions(t, want, &got)
			}
		})
	}
}

func TestSetShiurim(t *testing.T) {
	baseWant := func() *hebcal.CalOptions {
		return &hebcal.CalOptions{
			NumYears:           1,
			CandleLightingMins: 18,
		}
	}
	cases := []struct {
		Name  string
		Input []string
		Orig  *hebcal.CalOptions // default: new(hebcal.CalOptions)
		Want  *hebcal.CalOptions // if Err, ignore this
		Err   string
	}{
		{Name: "empty", Want: new(hebcal.CalOptions)},
		{Name: "base settings", Orig: baseWant(), Want: baseWant()},
		{
			Name:  "yerushalmi",
			Input: []string{"yerushalmi"},
			Orig:  baseWant(),
			Want: func() *hebcal.CalOptions {
				opts := baseWant()
				opts.YerushalmiYomi = true
				opts.YerushalmiEdition = yerushalmi.Vilna
				return opts
			}(),
		},
		{
			Name:  "yerushalmi:vilna",
			Input: []string{"yerushalmi:vilna"},
			Orig:  baseWant(),
			Want: func() *hebcal.CalOptions {
				opts := baseWant()
				opts.YerushalmiYomi = true
				opts.YerushalmiEdition = yerushalmi.Vilna
				return opts
			}(),
		},
		{
			Name:  "yerushalmi:schottenstein",
			Input: []string{"yerushalmi:schottenstein"},
			Orig:  baseWant(),
			Want: func() *hebcal.CalOptions {
				opts := baseWant()
				opts.YerushalmiYomi = true
				opts.YerushalmiEdition = yerushalmi.Schottenstein
				return opts
			}(),
		},
		{
			Name:  "mishna-yomi",
			Input: []string{"mishna-yomi"},
			Orig:  baseWant(),
			Want: func() *hebcal.CalOptions {
				opts := baseWant()
				opts.MishnaYomi = true
				return opts
			}(),
		},
		{
			Name:  "daf-yomi",
			Input: []string{"daf-yomi"},
			Orig:  baseWant(),
			Want: func() *hebcal.CalOptions {
				opts := baseWant()
				opts.DafYomi = true
				return opts
			}(),
		},
		{
			Name:  "nach-yomi",
			Input: []string{"nach-yomi"},
			Orig:  baseWant(),
			Want: func() *hebcal.CalOptions {
				opts := baseWant()
				opts.NachYomi = true
				return opts
			}(),
		},

		{
			Name: "multiple",
			// can't test all at once,
			// since yerushalmi and yerushalmi:schottenstein conflict
			Input: []string{
				"yerushalmi",
				"mishna-yomi",
				"daf-yomi",
				"nach-yomi",
			},
			Want: &hebcal.CalOptions{
				YerushalmiYomi:    true,
				YerushalmiEdition: yerushalmi.Vilna,
				MishnaYomi:        true,
				DafYomi:           true,
				NachYomi:          true,
			},
		},

		{
			Name:  "unknown",
			Input: []string{"unknown"},
			Err:   `unrecognized item(s) in shiurim: ["unknown"]`,
		},
		{
			Name:  "unknown and invalid",
			Input: []string{"unknown", "invalid"},
			Err:   `unrecognized item(s) in shiurim: ["unknown" "invalid"]`,
		},
		{
			Name:  "invalid empty string shiur",
			Input: []string{""},
			Err:   `unrecognized item(s) in shiurim: [""]`,
		},

		{
			Name: "conflict yerushalmi:vilna vs yerushalmi:schottenstein",
			Input: []string{
				"yerushalmi:vilna",
				"yerushalmi:schottenstein",
			},
			Err: "shiurim: conflicting yerushalmi edition settings found",
		},
		{
			Name: "conflict yerushalmi:schottenstein vs yerushalmi:vilna",
			Input: []string{
				"yerushalmi:schottenstein",
				"yerushalmi:vilna",
			},
			Err: "shiurim: conflicting yerushalmi edition settings found",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Orig == nil {
				c.Orig = new(hebcal.CalOptions)
			}
			if c.Want == nil {
				c.Want = new(hebcal.CalOptions)
			}
			got := *c.Orig // copy
			err := config.SetShiurim(&got, c.Input)
			test.CheckErr(t, err, c.Err)
			if c.Err == "" { // otherwise don't care
				test.CheckCalOptions(t, c.Want, &got)
			}
		})
	}
}

func TestConfig_Location(t *testing.T) {
	cases := []struct {
		Name string
		Cfg  config.Config
		Want *zmanim.Location
		Err  string
		Log  string
	}{
		{
			Name: "empty uses default city",
			Want: &zmanim.Location{
				Name:        "New York",
				CountryCode: "US",
				Latitude:    40.71427,
				Longitude:   -74.00597,
				TimeZoneId:  "America/New_York",
			},
		},
		{
			Name: "custom timezone",
			Cfg:  config.Config{Timezone: "Asia/Jerusalem"},
			Want: &zmanim.Location{
				Name:        "New York (times in timezone Asia/Jerusalem)",
				CountryCode: "US",
				Latitude:    40.71427,
				Longitude:   -74.00597,
				TimeZoneId:  "Asia/Jerusalem",
			},
		},
		{
			Name: "named city",
			Cfg:  config.Config{City: "Denver"},
			Want: &zmanim.Location{
				Name:        "Denver",
				CountryCode: "US",
				Latitude:    39.73915,
				Longitude:   -104.9847,
				TimeZoneId:  "America/Denver",
			},
		},

		// Geo
		{
			Name: "unnamed Geo",
			Cfg: config.Config{
				Timezone: "UTC",
				Geo:      &config.Coordinates{1.5, 2.5},
			},
			Want: &zmanim.Location{
				Name:        "User Defined City",
				CountryCode: "ZZ",
				Latitude:    1.5,
				Longitude:   2.5,
				TimeZoneId:  "UTC",
			},
		},
		{
			Name: "named Geo",
			Cfg: config.Config{
				City:     "Global Origin",
				Timezone: "UTC",
				Geo:      &config.Coordinates{0, 0},
			},
			Want: &zmanim.Location{
				Name:        "Global Origin",
				CountryCode: "ZZ",
				Latitude:    0,
				Longitude:   0,
				TimeZoneId:  "UTC",
			},
		},
		{
			Name: "named Geo in Israel",
			Cfg: config.Config{
				City:     "Kotel",
				Timezone: "Asia/Jerusalem",
				Geo:      &config.Coordinates{31.7767, 25.2345},
				IL:       true,
			},
			Want: &zmanim.Location{
				Name:        "Kotel",
				CountryCode: "IL",
				Latitude:    31.7767,
				Longitude:   25.2345,
				TimeZoneId:  "Asia/Jerusalem",
			},
		},

		// Errors
		{
			Name: "invalid timezone",
			Cfg:  config.Config{Timezone: "INVALID"},
			Err:  "unknown time zone INVALID",
		},
		{
			Name: "geo with missing timezone",
			Cfg:  config.Config{Geo: new(config.Coordinates)},
			Err:  "geo is set, but timezone is missing",
		},
		{
			Name: "geo out of bounds",
			Cfg: config.Config{
				Geo:      &config.Coordinates{91.0, 0},
				Timezone: "UTC",
			},
			Err: "invalid geo: invalid latitude: 91.000000",
		},
		{
			Name: "unknown city",
			Cfg:  config.Config{City: "Unknown"},
			Err:  `unknown city: "Unknown"`,
			Log: strings.TrimSpace(`
unknown city: "Unknown"
Use a nearby city; or add geo.lat, geo.lon, and timezone.
To show available cities, run:
  hebcalfmt --info cities
`),
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			logBuf := test.Logger(t)
			got, err := c.Cfg.Location()
			test.CheckErr(t, err, c.Err)
			if c.Err == "" { // otherwise don't care
				if *c.Want != *got {
					t.Errorf("want:\n%#v\ngot:\n%#v", c.Want, got)
				}
			}
			if c.Log != "" && c.Log != strings.TrimSpace(logBuf.String()) {
				t.Errorf("want logs:\n%s\ngot:\n%s", c.Log, logBuf.String())
			}
		})
	}
}

func TestSetToday(t *testing.T) {
	want := hebcal.CalOptions{
		AddHebrewDates: true,
		Omer:           true,
		IsHebrewYear:   false,
	}
	var got hebcal.CalOptions
	config.SetToday(&got)
	test.CheckCalOptions(t, &want, &got)
}

func TestSetChagOnly(t *testing.T) {
	want := hebcal.CalOptions{
		Mask: event.CHAG | event.LIGHT_CANDLES |
			event.LIGHT_CANDLES_TZEIS | event.YOM_TOV_ENDS,
	}
	var got hebcal.CalOptions
	config.SetChagOnly(&got)
	test.CheckCalOptions(t, &want, &got)
}

func TestParseFile(t *testing.T) {
	file := func(s string) *fstest.MapFile {
		return &fstest.MapFile{Data: []byte(s)}
	}
	files := fstest.MapFS{
		"empty.txt":       file(""),
		"lineNumbers.txt": file("1\n2\n3"),
		"test.json":       file(`{"status": "ok"}`),
		"forceError.txt":  file("force error"),
	}
	splitLines := func(r io.Reader, fpath string) ([]string, error) {
		buf, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		if string(buf) == "force error" {
			return nil, errors.New("forced error")
		}
		return strings.Split(string(buf), "\n"), nil
	}

	cases := []struct {
		Fpath string
		Want  []string
		Err   string
	}{
		{Fpath: "", Err: "ParseFile: missing file path"},
		{Fpath: "empty.txt", Want: []string{""}},
		{Fpath: "forceError.txt", Err: "forced error"},
		{Fpath: "test.json", Want: []string{`{"status": "ok"}`}},
		{Fpath: "lineNumbers.txt", Want: strings.Fields("1 2 3")},
		{
			Fpath: "nonexistent.txt",
			Err:   `open nonexistent.txt: file does not exist`,
		},
	}
	for _, c := range cases {
		t.Run(c.Fpath, func(t *testing.T) {
			test.Slogger(t)
			var got []string
			err := config.ParseFile(files, c.Fpath, splitLines, &got)
			test.CheckErr(t, err, c.Err)
			if c.Err == "" && !slices.Equal(c.Want, got) {
				t.Errorf("want: %q\ngot:  %q", c.Want, got)
			}
		})
	}
}
