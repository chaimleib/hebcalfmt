package daterange_test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/hebcal/greg"
	"github.com/hebcal/hdate"

	"github.com/chaimleib/hebcalfmt/daterange"
	"github.com/chaimleib/hebcalfmt/test"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// hdatesEqual checks to see if the Day, Month and Year all match.
// It is needed, because the struct also caches the Rata Die date,
// and that field may or may not be populated.
func hdatesEqual(a, b hdate.HDate) bool {
	return a.Day() == b.Day() && a.Month() == b.Month() && a.Year() == b.Year()
}

func TestRangeType_String(t *testing.T) {
	cases := []struct {
		Input daterange.RangeType
		Want  string
	}{
		{Input: daterange.RangeTypeYear, Want: "YEAR"},
		{Input: daterange.RangeTypeMonth, Want: "MONTH"},
		{Input: daterange.RangeTypeToday, Want: "TODAY"},
		{Input: daterange.RangeTypeDay, Want: "DAY"},
		{Input: 99, Want: "UNKNOWN(99)"},
	}
	for _, c := range cases {
		t.Run(c.Want, func(t *testing.T) {
			got := c.Input.String()
			if c.Want != got {
				t.Errorf("want: %q, got: %q", c.Want, got)
			}
		})
	}
}

func TestSource_IsZero(t *testing.T) {
	cases := []struct {
		Name  string
		Input daterange.Source
		Want  bool
	}{
		{Name: "empty", Input: daterange.Source{}, Want: true},
		{
			Name: "empty with IsHebrewDate",
			// IsHebrewDate means something touched us.
			// But either FromTime or Now should have been set as well.
			// This is in an inconsistent state.
			Input: daterange.Source{IsHebrewDate: true},
			Want:  true,
		},
		{
			Name: "with Now",
			// A non-zero Now is evidence that we were created with FromArgs()..
			Input: daterange.Source{Now: date(2025, time.January, 1)},
			Want:  false,
		},
		{
			Name:  "with FromTime",
			Input: daterange.Source{FromTime: date(2025, time.January, 1)},
			Want:  false,
		},
		{
			Name: "empty with Args",
			// Args means a user provided a daterange spec as a slice of strings.
			// But if so, and we should have been initialized by FromArgs,
			// and Now should have been set as well.
			// This is in an inconsistent state.
			Input: daterange.Source{Args: []string{"2025"}},
			Want:  true,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := c.Input.IsZero()
			if c.Want != got {
				t.Errorf("unexpected IsZero for %#v: %v", c.Input, got)
			}
		})
	}
}

func TestDateRange_FromTime(t *testing.T) {
	cases := []struct {
		Name  string
		Input time.Time
		Want  daterange.DateRange
	}{
		{
			Name:  "ok",
			Input: date(2025, time.May, 2),
			Want: daterange.DateRange{
				Source:    daterange.Source{FromTime: date(2025, time.May, 2)},
				RangeType: daterange.RangeTypeDay,
				Day:       2,
				GregMonth: time.May,
				Year:      2025,
			},
		},
		{
			Name:  "zero time",
			Input: time.Time{},
			Want: daterange.DateRange{
				RangeType: daterange.RangeTypeDay,
				Day:       1,
				GregMonth: time.January,
				Year:      1,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := daterange.FromTime(c.Input)
			if got == nil {
				t.Errorf("got unexpected nil from daterange.FromTime()")
				return
			}
			test.CheckDateRange(t, c.Want, *got)
		})
	}
}

func TestDateRange_FromArgs(t *testing.T) {
	now := time.Date(2025, time.December, 30, 18, 51, 58, 932, time.UTC)
	cases := []struct {
		Name         string
		Args         []string
		IsHebrewDate bool
		Want         daterange.DateRange
		Err          string
	}{
		{
			Name: "empty args implies current year",
			Args: nil,
			Want: daterange.DateRange{
				Source:    daterange.Source{Args: nil, Now: now},
				RangeType: daterange.RangeTypeYear,
				Day:       0,
				GregMonth: 0,
				Year:      now.Year(),
			},
		},
		{
			Name:         "IsHebrewYear with empty args implies current Hebrew year",
			Args:         nil,
			IsHebrewDate: true,
			Want: daterange.DateRange{
				Source: daterange.Source{
					Args:         nil,
					IsHebrewDate: true,
					Now:          now,
				},
				RangeType:    daterange.RangeTypeYear,
				IsHebrewDate: true,
				Day:          0,
				GregMonth:    0,
				Year:         5786,
			},
		},
		{
			Name: "Gregorian year",
			Args: []string{"2024"},
			Want: daterange.DateRange{
				Source:    daterange.Source{Args: []string{"2024"}, Now: now},
				RangeType: daterange.RangeTypeYear,
				Day:       0,
				GregMonth: 0,
				Year:      2024,
			},
		},
		{
			Name:         "Hebrew year",
			Args:         []string{"5784"},
			IsHebrewDate: true,
			Want: daterange.DateRange{
				Source: daterange.Source{
					Args:         []string{"5784"},
					IsHebrewDate: true,
					Now:          now,
				},
				RangeType:    daterange.RangeTypeYear,
				IsHebrewDate: true,
				Day:          0,
				GregMonth:    0,
				Year:         5784,
			},
		},
		{
			Name: "Gregorian month",
			Args: []string{"5", "2024"},
			Want: daterange.DateRange{
				Source:    daterange.Source{Args: []string{"5", "2024"}, Now: now},
				RangeType: daterange.RangeTypeMonth,
				Day:       0,
				GregMonth: time.May,
				Year:      2024,
			},
		},
		{
			Name: "invalid Gregorian month spec - bad year",
			Args: []string{"5", "INVALID"},
			Err:  `invalid year: strconv.Atoi: parsing "INVALID": invalid syntax`,
		},
		{
			Name: "invalid Gregorian month spec - bad month format",
			// hdate tries a bit too hard and turns "INVALIDMONTH" into "Iyyar",
			// because of the initial "I", so use initial "B" which is unknown.
			Args: []string{"BADMONTH", "2024"},
			Err:  `Gregorian months must be numeric, got "BADMONTH"`,
		},
		{
			Name: "invalid Gregorian month spec - zero month",
			Args: []string{"0", "2024"},
			Err:  `invalid month: 0`,
		},
		{
			Name: "invalid Gregorian month spec - month out of range",
			Args: []string{"13", "2024"},
			Err:  `invalid month: 13`,
		},
		{
			Name:         "Hebrew month",
			Args:         []string{"Iyar", "5784"},
			IsHebrewDate: true,
			Want: daterange.DateRange{
				Source: daterange.Source{
					Args:         []string{"Iyar", "5784"},
					IsHebrewDate: true,
					Now:          now,
				},
				RangeType:    daterange.RangeTypeMonth,
				IsHebrewDate: true,
				Day:          0,
				HebMonth:     hdate.Iyyar,
				Year:         5784,
			},
		},
		{
			Name:         "Hebrew month forces IsHebrewYear",
			Args:         []string{"Iyar", "5784"},
			IsHebrewDate: false,
			Want: daterange.DateRange{
				Source: daterange.Source{
					Args:         []string{"Iyar", "5784"},
					IsHebrewDate: false,
					Now:          now,
				},
				RangeType:    daterange.RangeTypeMonth,
				IsHebrewDate: true,
				Day:          0,
				HebMonth:     hdate.Iyyar,
				Year:         5784,
			},
		},
		{
			Name:         "corrects Adar II to I",
			Args:         []string{"Adar2", "5783"},
			IsHebrewDate: false,
			Want: daterange.DateRange{
				Source: daterange.Source{
					Args:         []string{"Adar2", "5783"},
					IsHebrewDate: false,
					Now:          now,
				},
				RangeType:    daterange.RangeTypeMonth,
				IsHebrewDate: true,
				Day:          0,
				HebMonth:     hdate.Adar1,
				Year:         5783,
			},
		},
		{
			Name:         "accepts AdarII in a leap year",
			Args:         []string{"AdarII", "5784"},
			IsHebrewDate: true,
			Want: daterange.DateRange{
				Source: daterange.Source{
					Args:         []string{"AdarII", "5784"},
					IsHebrewDate: true,
					Now:          now,
				},
				RangeType:    daterange.RangeTypeMonth,
				IsHebrewDate: true,
				Day:          0,
				HebMonth:     hdate.Adar2,
				Year:         5784,
			},
		},
		{
			Name:         "invalid Hebrew month - month as number",
			Args:         []string{"5", "5784"},
			IsHebrewDate: true,
			Err:          `expected Hebrew month name, got a number: 5`,
		},
		{
			Name:         "invalid Hebrew month - unknown month",
			Args:         []string{"BADMONTH", "5784"},
			IsHebrewDate: true,
			Err:          `unknown Hebrew month: "BADMONTH"`,
		},
		{
			Name: "Gregorian day",
			Args: []string{"5", "2", "2024"},
			Want: daterange.DateRange{
				Source:    daterange.Source{Args: []string{"5", "2", "2024"}, Now: now},
				RangeType: daterange.RangeTypeDay,
				Day:       2,
				GregMonth: time.May,
				Year:      2024,
			},
		},
		{
			Name: "invalid Gregorian day - invalid day format",
			Args: []string{"5", "INVALIDDAY", "2024"},
			Err:  `invalid day: strconv.Atoi: parsing "INVALIDDAY": invalid syntax`,
		},
		{
			Name: "invalid Gregorian day - zero day",
			Args: []string{"5", "0", "2024"},
			Err:  `invalid day for May 2024: 0`,
		},
		{
			Name: "invalid Gregorian day - day out of range",
			Args: []string{"5", "32", "2024"},
			Err:  `invalid day for May 2024: 32`,
		},
		{
			Name: "invalid Gregorian day - invalid year format",
			Args: []string{"5", "2", "INVALIDYEAR"},
			Err:  `invalid year: strconv.Atoi: parsing "INVALIDYEAR": invalid syntax`,
		},
		{
			Name: "invalid Gregorian day - invalid month format",
			Args: []string{"BADMONTH", "2", "2024"},
			Err:  `Gregorian months must be numeric, got "BADMONTH"`,
		},
		{
			Name: "Gregorian day - YYYY-MM-DD",
			Args: []string{"2024-05-02"},
			Want: daterange.DateRange{
				Source:    daterange.Source{Args: []string{"2024-05-02"}, Now: now},
				RangeType: daterange.RangeTypeDay,
				Day:       2,
				GregMonth: time.May,
				Year:      2024,
			},
		},
		{
			Name: "Gregorian day - YYYY-M-D",
			Args: []string{"2024-5-2"},
			Want: daterange.DateRange{
				Source:    daterange.Source{Args: []string{"2024-5-2"}, Now: now},
				RangeType: daterange.RangeTypeDay,
				Day:       2,
				GregMonth: time.May,
				Year:      2024,
			},
		},
		{
			Name: "invalid Gregorian day - YYYY-MM-DD",
			Args: []string{"2024-05-32"},
			Err:  `parsing time "2024-05-32": day out of range`,
		},
		{
			Name:         "Hebrew day",
			Args:         []string{"Iyar", "2", "5784"},
			IsHebrewDate: true,
			Want: daterange.DateRange{
				Source: daterange.Source{
					Args:         []string{"Iyar", "2", "5784"},
					IsHebrewDate: true,
					Now:          now,
				},
				RangeType:    daterange.RangeTypeDay,
				IsHebrewDate: true,
				Day:          2,
				HebMonth:     hdate.Iyyar,
				Year:         5784,
			},
		},
		{
			Name:         "invalid Hebrew day - zero day",
			Args:         []string{"Iyar", "0", "5784"},
			IsHebrewDate: true,
			Err:          `invalid day for Iyyar 5784: 0`,
		},
		{
			Name:         "invalid Hebrew day - day out of range",
			Args:         []string{"Iyar", "31", "5784"},
			IsHebrewDate: true,
			Err:          `invalid day for Iyyar 5784: 31`,
		},
		{
			Name: "invalid Args - too many args",
			Args: []string{"Iyar", "31", "5784", "INVALIDEXTRA"},
			Err:  `expected at most 3 args for date range spec, got 4`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := daterange.FromArgs(c.Args, c.IsHebrewDate, now)
			test.CheckErr(t, err, c.Err)
			if err != nil {
				if got != nil {
					t.Errorf(
						"non-nil error:\n%v\nbut got unexpected non-nil value from daterange.FromArgs():\n%#v",
						err,
						got,
					)
				}
				return
			}

			if got == nil {
				t.Errorf("got unexpected nil from daterange.FromArgs()")
				return
			}
			test.CheckDateRange(t, c.Want, *got)
		})
	}
}

func TestDateRange_String(t *testing.T) {
	now := time.Date(2025, time.May, 2, 18, 51, 58, 932, time.UTC)
	cases := []struct {
		Input     daterange.DateRange
		WantInner string
	}{
		// Years
		{
			Input: daterange.DateRange{
				Source: daterange.Source{
					Args:         nil,
					IsHebrewDate: true,
					Now:          now,
				},
				RangeType:    daterange.RangeTypeYear,
				IsHebrewDate: true,
				Year:         5785,
			},
			WantInner: "5785 (Hebrew)",
		},
		{
			Input: daterange.DateRange{
				Source:    daterange.Source{Args: nil, Now: now},
				RangeType: daterange.RangeTypeYear,
				Year:      2025,
			},
			WantInner: "2025",
		},
		// Months
		{
			Input: daterange.DateRange{
				Source:    daterange.Source{Args: []string{"5", "2025"}, Now: now},
				RangeType: daterange.RangeTypeMonth,
				GregMonth: time.May,
				Year:      2025,
			},
			WantInner: "May 2025",
		},
		{
			Input: daterange.DateRange{
				Source: daterange.Source{
					Args: []string{"Iyyar", "5786"},
					Now:  now,
				},
				RangeType:    daterange.RangeTypeMonth,
				IsHebrewDate: true,
				HebMonth:     hdate.Iyyar,
				Year:         5786,
			},
			WantInner: "Iyyar 5786",
		},
		// Gregorian day
		{
			Input: daterange.DateRange{
				Source:    daterange.Source{Args: []string{"2025-05-02"}, Now: now},
				RangeType: daterange.RangeTypeDay,
				Day:       2,
				GregMonth: time.May,
				Year:      2025,
			},
			WantInner: "2 May 2025",
		},
		{
			Input: daterange.DateRange{
				Source:    daterange.Source{Args: nil, Now: now},
				RangeType: daterange.RangeTypeToday,
				Day:       2,
				GregMonth: time.May,
				Year:      2025,
			},
			WantInner: "2 May 2025 --today",
		},
		// Hebrew day
		{
			Input: daterange.DateRange{
				Source:       daterange.Source{Args: nil, IsHebrewDate: true, Now: now},
				RangeType:    daterange.RangeTypeDay,
				IsHebrewDate: true,
				Day:          2,
				HebMonth:     hdate.Iyyar,
				Year:         5786,
			},
			WantInner: "2 Iyyar 5786",
		},
		{
			Input: daterange.DateRange{
				Source:       daterange.Source{Args: nil, IsHebrewDate: true, Now: now},
				RangeType:    daterange.RangeTypeToday,
				IsHebrewDate: true,
				Day:          2,
				HebMonth:     hdate.Iyyar,
				Year:         5786,
			},
			WantInner: "2 Iyyar 5786 --today",
		},
		// Boundary conditions
		{
			Input: daterange.DateRange{
				Source:    daterange.Source{Args: []string{"0"}, Now: now},
				RangeType: daterange.RangeTypeYear,
				Day:       0,
				GregMonth: time.Month(0),
				HebMonth:  hdate.HMonth(0),
				Year:      0,
			},
			WantInner: "0",
		},
		{
			Input:     daterange.DateRange{},
			WantInner: "empty",
		},
	}
	for _, c := range cases {
		t.Run(c.WantInner, func(t *testing.T) {
			want := fmt.Sprintf("DateRange<%s>", c.WantInner)
			got := c.Input.String()
			if want != got {
				t.Errorf("want:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}

func TestDateRange_Start(t *testing.T) {
	// calculate our expected dates

	// dates for RangeTypeDay
	d := date(2025, time.December, 30)
	hd := hdate.FromTime(d)

	// dates for RangeTypeMonth
	dMonth := date(2025, time.December, 1)
	hdMonth := hdate.FromTime(dMonth)
	hMonth := hdate.New(hdMonth.Year(), hdMonth.Month(), 1)

	// dates for RangeTypeYear
	dYear := date(2025, time.January, 1)
	hdYear := hdate.FromTime(dYear)
	hYear := hdate.New(hdYear.Year(), hdate.Tishrei, 1)

	cases := []struct {
		Name     string
		DR       daterange.DateRange
		NoJulian bool
		Want     hdate.HDate
	}{
		{
			Name: "basic",
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: d},
				RangeType: daterange.RangeTypeDay,
				Day:       d.Day(),
				GregMonth: d.Month(),
				Year:      d.Year(),
			},
			Want: hd,
		},
		{
			Name: "basic today",
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: d},
				RangeType: daterange.RangeTypeToday,
				Day:       d.Day(),
				GregMonth: d.Month(),
				Year:      d.Year(),
			},
			Want: hd,
		},
		{
			Name:     "noJulian",
			NoJulian: true,
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: d},
				RangeType: daterange.RangeTypeDay,
				Day:       d.Day(),
				GregMonth: d.Month(),
				Year:      d.Year(),
			},
			Want: hd,
		},
		{
			Name: "Hebrew",
			DR: daterange.DateRange{
				Source:       daterange.Source{IsHebrewDate: true, FromTime: d},
				RangeType:    daterange.RangeTypeDay,
				IsHebrewDate: true,
				Day:          hd.Day(),
				HebMonth:     hd.Month(),
				Year:         hd.Year(),
			},
			Want: hd,
		},
		{
			Name:     "Hebrew noJulian",
			NoJulian: true,
			DR: daterange.DateRange{
				Source:       daterange.Source{IsHebrewDate: true, FromTime: d},
				RangeType:    daterange.RangeTypeDay,
				IsHebrewDate: true,
				Day:          hd.Day(),
				HebMonth:     hd.Month(),
				Year:         hd.Year(),
			},
			Want: hd,
		},
		// Month
		{
			Name: "basic month",
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: dMonth},
				RangeType: daterange.RangeTypeMonth,
				GregMonth: dMonth.Month(),
				Year:      dMonth.Year(),
			},
			Want: hdMonth,
		},
		{
			Name: "basic Hebrew month",
			DR: daterange.DateRange{
				Source:       daterange.Source{IsHebrewDate: true, FromTime: dMonth},
				IsHebrewDate: true,
				RangeType:    daterange.RangeTypeMonth,
				HebMonth:     hMonth.Month(),
				Year:         hMonth.Year(),
			},
			Want: hMonth,
		},
		// Year
		{
			Name: "basic year",
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: dYear},
				RangeType: daterange.RangeTypeYear,
				Year:      dYear.Year(),
			},
			Want: hdYear,
		},
		{
			Name: "basic Hebrew year",
			DR: daterange.DateRange{
				Source:       daterange.Source{IsHebrewDate: true, FromTime: dYear},
				IsHebrewDate: true,
				RangeType:    daterange.RangeTypeYear,
				Year:         hYear.Year(),
			},
			Want: hYear,
		},
		{
			Name: "invalid RangeType",
			DR: daterange.DateRange{
				Source:    daterange.Source{IsHebrewDate: true, FromTime: dYear},
				RangeType: daterange.RangeType(-1),
			},
			Want: hdate.HDate{},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var slogHandler test.RecordHandler
			if c.Name == "invalid RangeType" {
				// capture and check expected error
				slog.SetDefault(slog.New(&slogHandler))
				defer func() {
					const wantMessage = "called Start on a DateRange with an unknown RangeType"
					for _, r := range slogHandler.Records() {
						if r.Message != wantMessage {
							t.Log(r) // pass other records through to the test log
							continue
						}
						if r.Level != slog.LevelError {
							t.Errorf(
								"found expected slog.Error message, but with wrong Level - want: %s, got %s",
								slog.LevelError,
								r.Level,
							)
						}
						return
					}
					t.Errorf("did not find expected slog.Error message: %q", wantMessage)
				}()
			}

			got := c.DR.Start(c.NoJulian)
			if !hdatesEqual(c.Want, got) {
				t.Errorf("want: %s\n got: %s", c.Want, got)
			}
		})
	}
}

func TestDateRange_End(t *testing.T) {
	// calculate our expected dates

	// dates for RangeTypeDay
	d := date(2025, time.May, 30)
	hd := hdate.FromTime(d)

	// dates for RangeTypeMonth
	dMonth := d
	// Hebrew date of end of Gregorian month
	hdMonth := hdate.FromTime(date(
		dMonth.Year(),
		dMonth.Month(),
		greg.DaysIn(dMonth.Month(), dMonth.Year()),
	))
	hStartMonth := hdate.FromTime(dMonth)
	// Hebrew date of end of Hebrew month
	hMonth := hdate.New(
		hStartMonth.Year(),
		hStartMonth.Month(),
		hdate.DaysInMonth(hStartMonth.Month(), hStartMonth.Year()),
	)

	// dates for RangeTypeYear
	dYear := dMonth
	hdYear := hdate.FromTime(date(dMonth.Year(), time.December, 31))
	hStartYear := hdate.FromTime(dYear)
	// Hebrew date of end of Hebrew year
	hYear := hdate.New(
		hStartYear.Year(),
		hdate.Elul,
		hdate.DaysInMonth(hdate.Elul, hStartYear.Year()),
	)

	cases := []struct {
		Name     string
		DR       daterange.DateRange
		NoJulian bool
		Want     hdate.HDate
	}{
		{
			Name: "basic",
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: d},
				RangeType: daterange.RangeTypeDay,
				Day:       d.Day(),
				GregMonth: d.Month(),
				Year:      d.Year(),
			},
			Want: hd,
		},
		{
			Name: "basic today",
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: d},
				RangeType: daterange.RangeTypeToday,
				Day:       d.Day(),
				GregMonth: d.Month(),
				Year:      d.Year(),
			},
			Want: hd,
		},
		{
			Name:     "noJulian",
			NoJulian: true,
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: d},
				RangeType: daterange.RangeTypeDay,
				Day:       d.Day(),
				GregMonth: d.Month(),
				Year:      d.Year(),
			},
			Want: hd,
		},
		{
			Name: "Hebrew",
			DR: daterange.DateRange{
				Source:       daterange.Source{IsHebrewDate: true, FromTime: d},
				RangeType:    daterange.RangeTypeDay,
				IsHebrewDate: true,
				Day:          hd.Day(),
				HebMonth:     hd.Month(),
				Year:         hd.Year(),
			},
			Want: hd,
		},
		{
			Name:     "Hebrew noJulian",
			NoJulian: true,
			DR: daterange.DateRange{
				Source:       daterange.Source{IsHebrewDate: true, FromTime: d},
				RangeType:    daterange.RangeTypeDay,
				IsHebrewDate: true,
				Day:          hd.Day(),
				HebMonth:     hd.Month(),
				Year:         hd.Year(),
			},
			Want: hd,
		},
		// Month
		{
			Name: "basic month",
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: dMonth},
				RangeType: daterange.RangeTypeMonth,
				GregMonth: dMonth.Month(),
				Year:      dMonth.Year(),
			},
			Want: hdMonth,
		},
		{
			Name: "basic Hebrew month",
			DR: daterange.DateRange{
				Source:       daterange.Source{IsHebrewDate: true, FromTime: dMonth},
				IsHebrewDate: true,
				RangeType:    daterange.RangeTypeMonth,
				HebMonth:     hMonth.Month(),
				Year:         hMonth.Year(),
			},
			Want: hMonth,
		},
		// Year
		{
			Name: "basic year",
			DR: daterange.DateRange{
				Source:    daterange.Source{FromTime: dYear},
				RangeType: daterange.RangeTypeYear,
				Year:      dYear.Year(),
			},
			Want: hdYear,
		},
		{
			Name: "basic Hebrew year",
			DR: daterange.DateRange{
				Source:       daterange.Source{IsHebrewDate: true, FromTime: dYear},
				IsHebrewDate: true,
				RangeType:    daterange.RangeTypeYear,
				Year:         hYear.Year(),
			},
			Want: hYear,
		},
		{
			Name: "invalid RangeType",
			DR: daterange.DateRange{
				Source:    daterange.Source{IsHebrewDate: true, FromTime: dYear},
				RangeType: daterange.RangeType(-1),
			},
			Want: hdate.HDate{},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var slogHandler test.RecordHandler
			if c.Name == "invalid RangeType" {
				// capture and check expected error
				slog.SetDefault(slog.New(&slogHandler))
				defer func() {
					const wantMessage = "called End on a DateRange with an unknown RangeType"
					for _, r := range slogHandler.Records() {
						if r.Message != wantMessage {
							t.Log(r) // pass other records through to the test log
							continue
						}
						if r.Level != slog.LevelError {
							t.Errorf(
								"found expected slog.Error message, but with wrong Level - want: %s, got %s",
								slog.LevelError,
								r.Level,
							)
						}
						return
					}
					t.Errorf("did not find expected slog.Error message: %q", wantMessage)
				}()
			}

			got := c.DR.End(c.NoJulian)
			if !hdatesEqual(c.Want, got) {
				t.Errorf("want: %s\n got: %s", c.Want, got)
			}
		})
	}
}
