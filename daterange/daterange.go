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

var UnreachableError = errors.New("unreachable")

type RangeType int

const (
	RangeTypeYear RangeType = iota
	RangeTypeMonth
	RangeTypeDay
	RangeTypeToday
)

func (t RangeType) String() string {
	switch t {
	case RangeTypeDay:
		return "DAY"
	case RangeTypeMonth:
		return "MONTH"
	case RangeTypeToday:
		return "TODAY"
	case RangeTypeYear:
		return "YEAR"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

type Source struct {
	Args         []string
	IsHebrewDate bool
	Now          time.Time
	FromTime     time.Time
}

func (s Source) IsZero() bool {
	return s.Now.IsZero() && s.FromTime.IsZero()
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
			FromTime: t,
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
// If isHebrewDate is false,
// it will change to true if a Hebrew month is specified.
// Otherwise, it will be respected and errors raised
// if an invalid date for isHebrewDate is provided.
//
// now is only used for the calendar date it contains,
// and only if args is length 0.
func FromArgs(
	args []string,
	isHebrewDate bool,
	now time.Time,
) (*DateRange, error) {
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
					UnreachableError, dr.HebMonth)
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
	if dr.Source.IsZero() {
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

// Start returns the first day of the DateRange,
// where DateRange is of RangeType Day, Today, or Month.
// It is an error to call Start when the DateRange is of RangeType Year.
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

// End returns the first day of the DateRange,
// where DateRange is of RangeType Day, Today, or Month.
// It is an error to call Start when the DateRange is of RangeType Year.
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
