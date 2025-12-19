package templating

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/omer"
	"github.com/hebcal/hebcal-go/zmanim"
)

func HebcalFuncs(opts *hebcal.CalOptions) map[string]any {
	return map[string]any{
		// hdate.HDate
		"hdatePartsEqual":             HDatePartsEqual,
		"hdateParse":                  HDateParse,
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
		"lookupCity":  zmanim.LookupCity,
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
		"hebcal": func() ([]event.CalEvent, error) {
			return hebcal.HebrewCalendar(opts)
		},

		// as<Type>Event converts [event.CalEvent]s to struct types.
		// It returns nil if it fails.
		"asHolidayEvent": AsEvent[event.HolidayEvent],
		"asOmerEvent":    AsEvent[omer.OmerEvent],
		"asTimedEvent":   AsEvent[hebcal.TimedEvent],
		"asUserEvent":    AsEvent[event.UserEvent],

		// timedEvents returns a slice of [hebcal.TimedEvent]
		"timedEvents": func(z *zmanim.Zmanim) ([]hebcal.TimedEvent, error) {
			return TimedEvents(opts, z)
		},

		// Modifying opts from a template
		"setStart":        SetStart(opts),
		"setEnd":          SetEnd(opts),
		"setYear":         SetYear(opts),
		"setNumYears":     SetNumYears(opts),
		"setIsHebrewYear": SetIsHebrewYear(opts),
	}
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

// TimedEvents uses the given opts and returns just the [event.CalEvent]s
// which are [hebcal.TimedEvent]s.
// This is an easy way to pull zmanim from Hebcal,
// if you don't want to fully customize your zmanim list yourself.
// This interface is very similar
// to the one provided by the classic hebcal binary.
//
// Unlike Hebcal, results are sorted by time, and certain ties are broken
// by putting Havdalah first and Candle lighting last.
func TimedEvents(
	opts *hebcal.CalOptions,
	z *zmanim.Zmanim,
) ([]hebcal.TimedEvent, error) {
	optsCopy := *opts
	opts = &optsCopy
	opts.Month = z.Month
	opts.Year = z.Year
	cal, err := hebcal.HebrewCalendar(opts)
	if err != nil {
		return nil, err
	}

	var results []hebcal.TimedEvent
	for _, evt := range cal {
		d := evt.GetDate().Gregorian()
		if d.Day() != z.Day {
			continue
		}
		timedEv, ok := evt.(hebcal.TimedEvent)
		if !ok {
			continue
		}
		results = append(results, timedEv)
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

// HDatePartsEqual compares the Hebrew year, month and day,
// and returns true if the two HDates match on all those fields.
func HDatePartsEqual(a, b hdate.HDate) bool {
	ay, am, ad := a.Day(), a.Month(), a.Year()
	by, bm, bd := b.Day(), b.Month(), b.Year()
	return ay == by && am == bm && ad == bd
}

// HDateParse parses a string in DD MMMM YYYY or MMMM DD YYYY format
// into an HDate.
// MMMM is the name of the Hebrew month, like "Adar II", "Adar 2", or "Tishrei".
func HDateParse(s string) (hdate.HDate, error) {
	var rv hdate.HDate

	parts := strings.Fields(s)
	length := len(parts)
	if length > 4 { // 4 in case of Adar I/II/1/2
		return rv, fmt.Errorf("too many words in a Hebrew date: %q", s)
	}
	if length < 3 {
		return rv, fmt.Errorf("too few words in a Hebrew date: %q", s)
	}

	year, err := strconv.Atoi(parts[length-1])
	if err != nil {
		return rv, fmt.
			Errorf("could not parse last word of Hebrew date as year: %q", s)
	}

	var monthParts []string
	day, err := strconv.Atoi(parts[0])
	if err != nil {
		day, err = strconv.Atoi(parts[length-2])
		if err != nil {
			return rv, fmt.Errorf("could not parse day from Hebrew date: %q", s)
		}
		monthParts = parts[0 : length-2]
	} else {
		monthParts = parts[1 : length-1]
	}

	monthStr := strings.Join(monthParts, " ")
	month, err := hdate.MonthFromName(monthStr)
	if err != nil {
		return rv, fmt.Errorf("failed to parse month from Hebrew date: %q", s)
	}

	return hdate.New(year, month, day), nil
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
