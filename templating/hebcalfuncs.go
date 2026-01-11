package templating

import (
	"fmt"
	"sort"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/omer"
	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/xhdate"
)

func HebcalFuncs(opts *hebcal.CalOptions) map[string]any {
	return map[string]any{
		// hdate.HDate
		"hdateEqual":                  xhdate.Equal,
		"hdateParse":                  xhdate.Parse,
		"hdateIsLeapYear":             hdate.IsLeapYear,
		"hdateMonthsInYear":           hdate.MonthsInYear,
		"hdateDaysInYear":             hdate.DaysInYear,
		"hdateLongCheshvan":           hdate.LongCheshvan,
		"hdateShortKislev":            hdate.ShortKislev,
		"hdateDaysInMonth":            hdate.DaysInMonth,
		"hdateToRD":                   hdate.ToRD,
		"hdateNew":                    hdate.New,
		"hdateFromRD":                 hdate.FromRD,
		"hdateFromGregorian":          hdate.FromGregorian,
		"hdateFromProlepticGregorian": hdate.FromProlepticGregorian,
		"hdateFromTime":               hdate.FromTime,
		"hdateMonthFromName":          hdate.MonthFromName,
		"hdateDayOnOrBefore":          hdate.DayOnOrBefore,

		// zmanim.Location
		"lookupCity":  LookupCity,
		"allCities":   zmanim.AllCities,
		"newLocation": zmanim.NewLocation,

		// zmanim.Zmanim
		"forDate": func(t time.Time) *zmanim.Zmanim {
			result := zmanim.New(opts.Location, t)
			return &result
		},
		"forLocationDate": func(loc *zmanim.Location, d time.Time) *zmanim.Zmanim {
			// Most consumers actually take a pointer, so convert it.
			result := zmanim.New(loc, d)
			return &result
		},

		// hebcal returns a slice of [event.CalEvent].
		// Underlying types of that interface can be recovered
		// using as<Kind>Event functions.
		"hebcal": Hebcal(opts),

		// as<Type>Event converts [event.CalEvent]s to struct types.
		// It returns nil if it fails.
		"asHolidayEvent": AsEvent[event.HolidayEvent],
		"asOmerEvent":    AsEvent[omer.OmerEvent],
		"asTimedEvent":   AsEvent[hebcal.TimedEvent],
		"asUserEvent":    AsEvent[event.UserEvent],

		// timedEvents returns a slice of [hebcal.TimedEvent]
		"timedEvents": TimedEvents(opts),

		// Modifying opts from a template
		"setDates":        SetDates(opts),
		"setStart":        SetStart(opts),
		"setEnd":          SetEnd(opts),
		"setYear":         SetYear(opts),
		"setNumYears":     SetNumYears(opts),
		"setIsHebrewYear": SetIsHebrewYear(opts),
	}
}

// LookupCity is the same as [zmanim.LookupCity],
// except that we return an error if no match is found.
func LookupCity(city string) (*zmanim.Location, error) {
	l := zmanim.LookupCity(city)
	if l == nil {
		return nil, fmt.Errorf("unknown city %q", city)
	}
	return l, nil
}

// AsEvent attempts to convert an [event.CalEvent] interface
// to its underlying concrete type.
// If the conversion fails, it returns nil.
func AsEvent[T event.CalEvent](e event.CalEvent) *T {
	te, ok := e.(T)
	if ok {
		return &te
	}
	return nil
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

// Hebcal returns a slice of event.CalEvent like the original hebcal program.
// If no dates are provided, it uses the hebcal.CalOptions
// to select a date range.
// If one date is provided, only the events for that day are returned.
// If two dates, all the events between them are returned,
// including those on the end date.
func Hebcal(
	opts *hebcal.CalOptions,
) func(dates ...hdate.HDate) ([]event.CalEvent, error) {
	return func(dates ...hdate.HDate) ([]event.CalEvent, error) {
		optsCopy := *opts
		opts := &optsCopy
		if _, err := SetDates(opts)(dates...); err != nil {
			return nil, err
		}
		return hebcal.HebrewCalendar(opts)
	}
}

// TimedEvents uses the given opts and returns just the [event.CalEvent]s
// which are [hebcal.TimedEvent]s.
// This is an easy way to pull zmanim from Hebcal,
// if you don't want to fully customize your zmanim list yourself.
// This interface is very similar
// to the one provided by the classic hebcal binary.
//
// Unlike Hebcal, results are sorted by time, and certain ties are broken
// by putting Havdalah first and Candle lighting last.
//
// If no dates are provided, it uses the hebcal.CalOptions
// to select a date range.
// If one date is provided, only the events for that day are returned.
// If two dates, all the events between them are returned,
// including those on the end date.
func TimedEvents(
	opts *hebcal.CalOptions,
) func(dates ...hdate.HDate) ([]hebcal.TimedEvent, error) {
	return func(dates ...hdate.HDate) ([]hebcal.TimedEvent, error) {
		optsCopy := *opts
		opts := &optsCopy
		if _, err := SetDates(opts)(dates...); err != nil {
			return nil, err
		}

		cal, err := hebcal.HebrewCalendar(opts)
		if err != nil {
			return nil, err
		}

		var results []hebcal.TimedEvent
		for _, evt := range cal {
			if timedEv, ok := evt.(hebcal.TimedEvent); ok {
				results = append(results, timedEv)
			}
		}

		sort.Slice(results, func(i, j int) bool {
			if results[i].EventTime.Equal(results[j].EventTime) {
				if results[i].Desc == "Havdalah" {
					return true
				} else if results[j].Desc == "Havdalah" {
					return false
				}
				if results[i].Desc == "Candle lighting" {
					return false
				} else if results[j].Desc == "Candle lighting" {
					return true
				}
			}
			return results[i].EventTime.Before(results[j].EventTime)
		})

		return results, nil
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
