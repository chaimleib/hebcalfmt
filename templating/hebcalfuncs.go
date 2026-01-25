package templating

import (
	"sort"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/omer"
)

func HebcalFuncs(opts *hebcal.CalOptions) map[string]any {
	return map[string]any{
		// as<Type>Event converts [event.CalEvent]s to struct types.
		// It returns nil if it fails.
		"asHolidayEvent": AsEvent[event.HolidayEvent],
		"asOmerEvent":    AsEvent[omer.OmerEvent],
		"asTimedEvent":   AsEvent[hebcal.TimedEvent],
		"asUserEvent":    AsEvent[event.UserEvent],

		// hebcal returns a slice of [event.CalEvent].
		// Underlying types of that interface can be recovered
		// using as<Kind>Event functions.
		"hebcal": Hebcal(opts),

		// timedEvents returns a slice of [hebcal.TimedEvent]
		"timedEvents":   TimedEvents(opts),
		"eventsByFlags": EventsByFlags,

		"dayHasFlags":          DayHasFlags(opts),
		"dayIsShabbatOrYomTov": DayIsShabbatOrYomTov(opts),
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

// MergeFlags combines the flags into a single mask.
func MergeFlags(flags ...event.HolidayFlags) event.HolidayFlags {
	var mask event.HolidayFlags
	for _, flag := range flags {
		mask |= flag
	}
	return mask
}

// EventsByFlags filters the events to just the ones
// which have any of the given flags set.
func EventsByFlags(
	events []event.CalEvent,
	flags ...event.HolidayFlags,
) []event.CalEvent {
	mask := MergeFlags(flags...)

	// Gather the matching events.
	var results []event.CalEvent
	for _, e := range events {
		if e.GetFlags()&mask != 0 {
			results = append(results, e)
		}
	}

	return results
}

// DayHasFlags returns whether d has any of the given flags set
// on any of its events.
func DayHasFlags(
	opts *hebcal.CalOptions,
) func(d hdate.HDate, flags ...event.HolidayFlags) (bool, error) {
	return func(d hdate.HDate, flags ...event.HolidayFlags) (bool, error) {
		mask := MergeFlags(flags...)

		// Get the events occurring on d.
		events, err := Hebcal(opts)(d)
		if err != nil {
			return false, err
		}

		// Check the events.
		for _, e := range events {
			if e.GetFlags()&mask != 0 {
				return true, nil
			}
		}
		return false, nil
	}
}

// DayIsShabbatOrYomTov returns whether melacha is forbidden on the given day.
// It may be used by logic which determines candle lighting and havdalah times.
func DayIsShabbatOrYomTov(
	opts *hebcal.CalOptions,
) func(d hdate.HDate) (bool, error) {
	return func(d hdate.HDate) (bool, error) {
		events, err := Hebcal(opts)(d)
		if err != nil {
			return false, err
		}
		for _, ev := range events {
			if ev.GetFlags()&event.CHAG != 0 {
				return true, nil
			}
		}
		return d.Weekday() == time.Saturday, nil
	}
}
