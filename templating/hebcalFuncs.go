package templating

import (
	"sort"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/zmanim"
	"github.com/nathan-osman/go-sunrise"
)

func HebcalFuncs(opts *hebcal.CalOptions) map[string]any {
	return map[string]any{
		// hdate.HDate
		"hdateFromTime":   hdate.FromTime,
		"hdatePartsEqual": HDatePartsEqual,

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
			result := zmanim.New(loc, d)
			return &result
		},

		// hebcal.TimedEvent
		"timedEvents": func(z *zmanim.Zmanim) ([]hebcal.TimedEvent, error) {
			return TimedEvents(opts, z)
		},

		// zmanim.Zmanim -> time.Time
		"timeAtAngle": TimeAtAngle,
		"hourOffset":  HourOffset,
	}
}

// HourOffset returns a time for the given number of halachic hours
// past sunrise.
func HourOffset(z *zmanim.Zmanim, tz *time.Location, hours float64) time.Time {
	rise := z.Sunrise()
	seconds := rise.Unix() + int64(z.Hour()*hours)
	return time.Unix(seconds, 0).In(tz)
}

// TimeAtAngle returns when the center of the sun
// will be at the given angle below the horizon.
// AM or PM can be chosen via the rising bool.
func TimeAtAngle(z *zmanim.Zmanim, tz *time.Location, angle float64, rising bool) time.Time {
	morning, evening := sunrise.TimeOfElevation(
		z.Location.Latitude,
		z.Location.Longitude,
		-angle,
		z.Year,
		z.Month,
		z.Day,
	)
	if rising {
		return InLoc(tz, morning)
	}
	return InLoc(tz, evening)
}

func InLoc(tz *time.Location, t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(tz)
}

// TimedEvents uses the given opts and returns just the [event.CalEvent]s
// which are [hebcal.TimedEvent]s.
// This is an easy way to pull zmanim from Hebcal,
// if you don't want to fully customize your zmanim list yourself.
// This interface is very similar
// to the one provided by the classic hebcal binary.
//
// Unlike Hebcal, results are sorted by time,
// and certain ties are broken by putting Havdalah first and Candle lighting last.
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
