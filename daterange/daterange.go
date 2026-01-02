package daterange

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hebcal/greg"
	"github.com/hebcal/hdate"
)

// ErrUnreachable means that there is a coding defect if returned.
var ErrUnreachable = errors.New("unreachable code")

// RangeType marks how long the requested DateRange should be.
// This is passed to classic hebcal to select calendar lengths
// of a year, a month, or a day.
type RangeType int

const (
	RangeTypeYear RangeType = iota
	RangeTypeMonth
	RangeTypeDay

	// RangeTypeToday appears to be identical in behavior to RangeTypeDay.
	// We provide it to interoperate with hebcal classic,
	// but its purpose is unknown.
	RangeTypeToday
)

func (t RangeType) String() string {
	switch t {
	case RangeTypeYear:
		return "YEAR"
	case RangeTypeMonth:
		return "MONTH"
	case RangeTypeDay:
		return "DAY"
	case RangeTypeToday:
		return "TODAY"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// Source describes how a [DateRange] was produced.
// If it was created with [WithArgs], the `Now` field will be set.
// If with [FromTime], the `FromTime` field will be non-nil.
type Source struct {
	Args         []string
	IsHebrewDate bool
	Now          time.Time
	FromTime     *time.Time
}

// Defaulted returns true if the user did not provide a date besides `Now`
// to set up the `DateRange`.
// This can happen if, when run from the CLI, no date was specified.
func (s Source) Defaulted() bool {
	return len(s.Args) == 0 && s.FromTime == nil
}

type DateRange struct {
	Source       Source
	RangeType    RangeType
	Day          int
	GregMonth    time.Month
	HebMonth     hdate.HMonth
	Year         int
	IsHebrewDate bool
}

// FromTime takes a time.Time and converts it to a single-day DateRange.
func FromTime(t time.Time) *DateRange {
	return &DateRange{
		Source: Source{
			FromTime: &t,
		},
		RangeType: RangeTypeDay,
		Day:       t.Day(),
		GregMonth: t.Month(),
		Year:      t.Year(),
	}
}

// FromArgs takes a slice of strings containing a date range spec,
// and returns the DateRange indicated.
//
// The `args` slice should be in the sequence `[[ month [ day ]] year]`,
// where `day` and `year` are numeric.
// `month` must be numeric for Gregorian months,
// or the name of a Hebrew month.
// For Adar 1 and 2, do not include any spaces.
//
// Even if `isHebrewDate` is false,
// the result's `IsHebrewDate` will be forced true
// if a Hebrew month is specified.
// Otherwise, it will be respected and errors raised
// if an invalid date for isHebrewDate is provided.
//
// If `args` is length 0, `now` is used for the calendar date it contains.
// Regardless, `now` must be non-zero, since it marks the DateRange as non-zero.
func FromArgs(
	args []string,
	isHebrewDate bool,
	now time.Time,
) (*DateRange, error) {
	if now.IsZero() {
		return nil, errors.New(
			"daterange.FromArgs: now must not be a zero time",
		)
	}

	dr := new(DateRange)
	dr.Source = Source{
		Args:         args,
		IsHebrewDate: isHebrewDate,
		Now:          now,
	}
	dr.IsHebrewDate = isHebrewDate

	// TrimSpace all args
	args = slices.Clone(args)
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}

	switch len(args) {
	case 0:
		if isHebrewDate {
			hd := hdate.FromGregorian(now.Year(), now.Month(), now.Day())
			dr.Year = hd.Year()
		} else {
			dr.Year = now.Year()
		}

	case 1:
		arg0 := args[0]
		yy, err := strconv.Atoi(arg0)
		if err == nil {
			dr.Year = yy /* just year specified */
			break
		}

		// Use custom date format,
		// since time.DateOnly requires leading zeroes for month and day.
		t, err := time.Parse("2006-1-2", arg0)
		if err != nil {
			return nil, err
		}

		dr.Year = t.Year()
		dr.GregMonth = t.Month()
		dr.Day = t.Day()
		dr.RangeType = RangeTypeDay

	case 2:
		yy, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("invalid year: %w", err)
		}
		dr.Year = yy
		if err := dr.parseGregOrHebMonth(args[0]); err != nil {
			return nil, err
		}
		dr.RangeType = RangeTypeMonth

	case 3:
		dd, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("invalid day: %w", err)
		}
		dr.Day = dd

		yy, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, fmt.Errorf("invalid year: %w", err)
		}
		dr.Year = yy

		if err := dr.parseGregOrHebMonth(args[0]); err != nil {
			return nil, err
		}

		dr.RangeType = RangeTypeDay

	default:
		return nil, fmt.
			Errorf("expected at most 3 args for date range spec, got %d", len(args))
	}

	// Check months
	switch dr.RangeType {
	case RangeTypeMonth, RangeTypeDay, RangeTypeToday:
		if dr.IsHebrewDate {
			// invalid Hebrew month ranges should be impossible
			lastMonth := hdate.HMonth(hdate.MonthsInYear(dr.Year))
			if dr.HebMonth <= 0 || dr.HebMonth > lastMonth {
				slog.Error("impossible Hebrew month",
					"HebMonth", dr.HebMonth,
					"daterange", dr,
				)
				return nil, fmt.Errorf("%w: invalid month: %v",
					ErrUnreachable, dr.HebMonth)
			}
		} else {
			const lastMonth = time.December
			if dr.GregMonth <= 0 || dr.GregMonth > lastMonth {
				return nil, fmt.Errorf("invalid month: %d",
					dr.GregMonth)
			}
		}
	}

	// Check days in month
	if dr.RangeType == RangeTypeDay {
		if dr.IsHebrewDate {
			lastDay := hdate.DaysInMonth(dr.HebMonth, dr.Year)
			if dr.Day <= 0 || dr.Day > lastDay {
				return nil, fmt.Errorf("invalid day for %s %d: %d",
					dr.HebMonth, dr.Year, dr.Day)
			}
		} else {
			lastDay := greg.DaysIn(dr.GregMonth, dr.Year)
			if dr.Day <= 0 || dr.Day > lastDay {
				return nil, fmt.Errorf("invalid day for %s %d: %d",
					dr.GregMonth, dr.Year, dr.Day)
			}
		}
	}
	return dr, nil
}

func (dr *DateRange) parseGregOrHebMonth(arg string) (err error) {
	dr.IsHebrewDate, dr.GregMonth, dr.HebMonth, err = parseGregOrHebMonth(
		dr.IsHebrewDate, dr.Year, arg)
	return
}

func parseGregOrHebMonth(
	isHebrewYear bool,
	theYear int,
	arg string,
) (
	newIsHebrewYear bool,
	gregMonth time.Month,
	hebMonth hdate.HMonth,
	err error,
) {
	mm, err := strconv.Atoi(arg)
	if err == nil {
		if isHebrewYear {
			err = fmt.Errorf("expected Hebrew month name, got a number: %v", mm)
			return
		}
		gregMonth = time.Month(mm) /* gregorian month */
		return
	}

	hm, err := hdate.MonthFromName(arg)
	if err != nil {
		if isHebrewYear {
			err = fmt.Errorf("unknown Hebrew month: %q", arg)
		} else {
			err = fmt.Errorf("Gregorian months must be numeric, got %q", arg)
		}
		return
	}

	hebMonth = hm
	newIsHebrewYear = true /* automagically turn it on */
	if hm == hdate.Adar2 && !hdate.IsLeapYear(theYear) {
		hebMonth = hdate.Adar1 /* silently fix this mistake */
	}
	return
}

func (dr DateRange) String() string {
	return fmt.Sprintf("DateRange<%s>", dr.basicString())
}

func (dr DateRange) basicString() string {
	if dr.Source.Now.IsZero() && dr.Source.FromTime == nil {
		return "empty"
	}

	switch dr.RangeType {
	case RangeTypeYear:
		if dr.IsHebrewDate {
			return fmt.Sprintf("%d (Hebrew)", dr.Year)
		}
		return strconv.Itoa(dr.Year)

	case RangeTypeMonth:
		if dr.IsHebrewDate {
			return fmt.Sprintf("%s %d", dr.HebMonth, dr.Year)
		}
		return fmt.Sprintf("%s %d", dr.GregMonth, dr.Year)

	default:
		var t string
		if dr.RangeType == RangeTypeToday {
			t = " --today"
		}

		if dr.IsHebrewDate {
			return fmt.Sprintf("%d %s %d%s", dr.Day, dr.HebMonth, dr.Year, t)
		}
		return fmt.Sprintf("%d %s %d%s", dr.Day, dr.GregMonth, dr.Year, t)
	}
}

func fromGregorianFunc(
	noJulian bool,
) func(y int, m time.Month, d int) hdate.HDate {
	if noJulian {
		return hdate.FromProlepticGregorian
	}
	return hdate.FromGregorian
}

// Start returns the first day of the DateRange.
func (dr DateRange) Start(noJulian bool) hdate.HDate {
	fromGregorian := fromGregorianFunc(noJulian)

	switch dr.RangeType {
	case RangeTypeToday, RangeTypeDay:
		if dr.IsHebrewDate {
			return hdate.New(dr.Year, dr.HebMonth, dr.Day)
		}
		return fromGregorian(dr.Year, dr.GregMonth, dr.Day)

	case RangeTypeMonth:
		if dr.IsHebrewDate {
			return hdate.New(dr.Year, dr.HebMonth, 1)
		}
		return fromGregorian(dr.Year, dr.GregMonth, 1)

	case RangeTypeYear:
		if dr.IsHebrewDate {
			return hdate.New(dr.Year, hdate.Tishrei, 1)
		}
		return fromGregorian(dr.Year, time.January, 1)

	default:
		slog.Error(
			"called Start on a DateRange with an unknown RangeType",
			"rangeType", dr.RangeType.String(),
			"dateRange", dr.String(),
			"noJulian", noJulian,
		)
		return hdate.HDate{}
	}
}

// End returns the last day of the DateRange.
func (dr DateRange) End(noJulian bool) hdate.HDate {
	fromGregorian := fromGregorianFunc(noJulian)

	switch dr.RangeType {
	case RangeTypeToday, RangeTypeDay:
		if dr.IsHebrewDate {
			return hdate.New(dr.Year, dr.HebMonth, dr.Day)
		}
		return fromGregorian(dr.Year, dr.GregMonth, dr.Day)

	case RangeTypeMonth:
		if dr.IsHebrewDate {
			return hdate.New(dr.Year, dr.HebMonth, hdate.DaysInMonth(
				dr.HebMonth,
				dr.Year,
			))
		}
		return fromGregorian(dr.Year, dr.GregMonth, greg.DaysIn(
			dr.GregMonth,
			dr.Year,
		))

	case RangeTypeYear:
		if dr.IsHebrewDate {
			return hdate.New(dr.Year+1, hdate.Tishrei, 1).Prev()
		}
		return fromGregorian(dr.Year, time.December, 31)

	default:
		slog.Error(
			"called End on a DateRange with an unknown RangeType",
			"rangeType", dr.RangeType.String(),
			"dateRange", dr.String(),
			"noJulian", noJulian,
		)
		return hdate.HDate{}
	}
}

// StartOrToday returns the first day of the DateRange,
// like [(DateRange).Start], unless [(Source).Defaulted].
// Where that method defaults to the current year, this one defaults to today.
func (dr DateRange) StartOrToday(noJulian bool) hdate.HDate {
	if dr.Source.Defaulted() {
		fromGregorian := fromGregorianFunc(noJulian)
		now := dr.Source.Now
		return fromGregorian(now.Year(), now.Month(), now.Day())
	}
	return dr.Start(noJulian)
}
