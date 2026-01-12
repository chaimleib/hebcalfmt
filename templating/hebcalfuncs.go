package templating

import (
	"sort"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/omer"
)

func HebcalFuncs(opts *hebcal.CalOptions) map[string]any {
	return map[string]any{
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

// CompareTimedEvents allows hebcal.TimedEvents to be sorted.
// When clock times match, ties are broken by prioritizing
// Havdalah first and Candle lighting last.
// For example:
//
//	sort.Slice(events, CompareTimedEvents(events))
func CompareTimedEvents(events []hebcal.TimedEvent) func(i, j int) bool {
	return func(i, j int) bool {
		if events[i].EventTime.Equal(events[j].EventTime) {
			if events[i].Desc == "Havdalah" {
				return true
			} else if events[j].Desc == "Havdalah" {
				return false
			}
			if events[i].Desc == "Candle lighting" {
				return false
			} else if events[j].Desc == "Candle lighting" {
				return true
			}
		}
		return events[i].EventTime.Before(events[j].EventTime)
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

		sort.Slice(results, CompareTimedEvents(results))
		return results, nil
	}
}
