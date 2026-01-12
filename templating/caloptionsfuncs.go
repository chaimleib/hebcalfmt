package templating

import (
	"fmt"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/hebcal"
)

// CalOptionsFuncs contains functions for Modifying opts from a template.
func CalOptionsFuncs(opts *hebcal.CalOptions) map[string]any {
	return map[string]any{
		"setDates":        SetDates(opts),
		"setStart":        SetStart(opts),
		"setEnd":          SetEnd(opts),
		"setYear":         SetYear(opts),
		"setNumYears":     SetNumYears(opts),
		"setIsHebrewYear": SetIsHebrewYear(opts),
	}
}

// SetDates configures the opts with the provided dates.
// If no dates are provided, it does nothing.
// If one date is provided, it sets both opts.Start and opts.End
// to select that single date.
// If two dates, it checks that the first date is before the second,
// and sets opts.Start with the first and opts.End with the second.
// If more dates are provided, we return an error.
func SetDates(opts *hebcal.CalOptions) func(dates ...hdate.HDate) (any, error) {
	return func(dates ...hdate.HDate) (any, error) {
		switch len(dates) {
		case 0:
			// pass

		case 1:
			SetStart(opts)(dates[0])
			SetEnd(opts)(dates[0])

		case 2:
			start, startGreg := dates[0], dates[0].Gregorian()
			end, endGreg := dates[1], dates[1].Gregorian()
			if startGreg.After(endGreg) {
				return "", fmt.Errorf(
					"first date must be before second, got %s/%s, %s/%s",
					start, startGreg.Format(time.DateOnly),
					end, endGreg.Format(time.DateOnly),
				)
			}
			SetStart(opts)(start)
			SetEnd(opts)(end)

		default:
			return "", fmt.Errorf("expected 0-2 dates, got %d", len(dates))
		}

		return "", nil
	}
}

// SetStart tells template functions like hebcal and timedEvents the date
// from which events should be returned.
func SetStart(opts *hebcal.CalOptions) func(hd hdate.HDate) any {
	return func(hd hdate.HDate) any {
		opts.NumYears = 1
		opts.Year = 0
		opts.Start = hd
		return ""
	}
}

// SetEnd tells template functions like hebcal and timedEvents the date
// through which events should be returned.
func SetEnd(opts *hebcal.CalOptions) func(hd hdate.HDate) any {
	return func(hd hdate.HDate) any {
		opts.NumYears = 1
		opts.Year = 0
		opts.End = hd
		return ""
	}
}

// SetYear tells template functions like hebcal and timedEvents
// the starting year from which events should be returned.
func SetYear(opts *hebcal.CalOptions) func(y int) any {
	return func(y int) any {
		opts.Year = y
		opts.Start = hdate.HDate{}
		opts.End = hdate.HDate{}
		return ""
	}
}

// SetNumYears tells template functions like hebcal and timedEvents
// how many years of events should be returned.
func SetNumYears(opts *hebcal.CalOptions) func(n int) any {
	return func(n int) any {
		opts.NumYears = n
		return ""
	}
}

// SetIsHebrewYear tells template functions like hebcal and timedEvents
// how to interpret the Year and NumYears fields.
func SetIsHebrewYear(opts *hebcal.CalOptions) func(b bool) any {
	return func(b bool) any {
		opts.IsHebrewYear = b
		return ""
	}
}
