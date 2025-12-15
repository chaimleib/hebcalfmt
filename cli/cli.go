package cli

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/zmanim"
	"github.com/nathan-osman/go-sunrise"
)

func Run() int {
	log.SetFlags(0)

	opts, tmpl, err := handleArgs()
	if err != nil {
		log.Println(err)
		return 1
	}

	now := time.Now()
	z := zmanim.New(opts.Location, now)

	tz, err := time.LoadLocation(opts.Location.TimeZoneId)
	if err != nil {
		log.Println(err)
		return 1
	}

	err = tmpl.Execute(os.Stdout, map[string]any{
		"now":      now,
		"tz":       tz,
		"location": opts.Location,
		"z":        &z,
		"time": map[string]any{
			"Hour":        time.Hour,
			"Minute":      time.Minute,
			"Second":      time.Second,
			"Millisecond": time.Millisecond,
			"Microsecond": time.Microsecond,
			"Nanosecond":  time.Nanosecond,

			"Layout":      time.Layout,
			"ANSIC":       time.ANSIC,
			"UnixDate":    time.UnixDate,
			"RubyDate":    time.RubyDate,
			"RFC822":      time.RFC822,
			"RFC822Z":     time.RFC822Z,
			"RFC850":      time.RFC850,
			"RFC1123":     time.RFC822,
			"RFC1123Z":    time.RFC822Z,
			"RFC3339":     time.RFC3339,
			"RFC3339Nano": time.RFC3339Nano,
			"Kitchen":     time.Kitchen,

			"Stamp":      time.Stamp,
			"StampMilli": time.StampMilli,
			"StampMicro": time.StampMicro,
			"StampNano":  time.StampNano,
			"DateTime":   time.DateTime,
			"DateOnly":   time.DateOnly,
			"TimeOnly":   time.TimeOnly,

			"January":   time.January,
			"February":  time.February,
			"March":     time.March,
			"April":     time.April,
			"May":       time.May,
			"June":      time.June,
			"July":      time.July,
			"August":    time.August,
			"September": time.September,
			"October":   time.October,
			"November":  time.November,
			"December":  time.December,

			"Sunday":    time.Sunday,
			"Monday":    time.Monday,
			"Tuesday":   time.Tuesday,
			"Wednesday": time.Wednesday,
			"Thursday":  time.Thursday,
			"Friday":    time.Friday,
			"Saturday":  time.Saturday,
		},
		"event": map[string]any{
			"ZMANIM": event.ZMANIM,
		},
	})
	if err != nil {
		log.Println(err)
		return 1
	}

	return 0
}

func handleArgs() (opts *hebcal.CalOptions, tmpl *template.Template, err error) {
	if len(os.Args) != 2 {
		log.Println(usage())
		return nil, nil, fmt.Errorf("expected path to a template")
	}

	tmpl = new(template.Template)
	tmpl = tmpl.Funcs(map[string]any{
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

		// time.Duration
		"secondsDuration": func(secs float64) time.Duration {
			return time.Duration(secs * 1_000_000_000)
		},
		"timeParseDuration": time.ParseDuration,
		"timeSince":         time.Since,
		"timeUntil":         time.Until,

		// time.Location
		"timeFixedZone":    time.FixedZone,
		"timeLoadLocation": time.LoadLocation,

		// time.Time
		"timeParse":           time.Parse,
		"timeParseInLocation": time.ParseInLocation,
		"timeUnix":            time.Unix,
		"timeUnixMicro":       time.UnixMicro,
		"timeUnixMilli":       time.UnixMilli,
		"timeDate":            time.Date,
		"datePartsEqual":      DatePartsEqual,

		// strings
		"contains":        strings.Contains,
		"containsAny":     strings.ContainsAny,
		"count":           strings.Count,
		"equalFold":       strings.EqualFold,
		"hasPrefix":       strings.HasPrefix,
		"hasSuffix":       strings.HasSuffix,
		"stringsIndex":    strings.Index,
		"stringsIndexAny": strings.IndexAny,
		"join":            strings.Join,
		"lastIndex":       strings.LastIndex,
		"lastIndexAny":    strings.LastIndexAny,
		"lines":           strings.Lines,
		"repeat":          strings.Repeat,
		"replace":         strings.Replace,
		"replaceAll":      strings.ReplaceAll,
		"split":           strings.Split,
		"splitAfter":      strings.SplitAfter,
		"splitAfterN":     strings.SplitAfterN,
		"splitAfterSeq":   strings.SplitAfterSeq,
		"splitN":          strings.SplitN,
		"splitSeq":        strings.SplitSeq,
		"toLower":         strings.ToLower,
		"toTitle":         strings.ToTitle,
		"toUpper":         strings.ToUpper,
		"toValidUTF8":     strings.ToValidUTF8,
		"trim":            strings.Trim,
		"trimLeft":        strings.TrimLeft,
		"tripPrefix":      strings.TrimPrefix,
		"trimRight":       strings.TrimRight,
		"trimSpace":       strings.TrimSpace,
		"trimSuffix":      strings.TrimSuffix,

		// type conversions
		"itof": func(i int) float64 { return float64(i) },
		"ftoi": func(f float64) int { return int(f) },

		// env
		"getenv": os.Getenv,
	})

	tmpl, err = parseFile(tmpl, os.Args[1])
	if err != nil {
		return opts, nil, err
	}

	loc := zmanim.LookupCity("Phoenix")
	now := time.Now()
	opts = &hebcal.CalOptions{
		Year:             now.Year(),
		Month:            now.Month(),
		Sedrot:           true,
		CandleLighting:   true,
		DailyZmanim:      true,
		Location:         loc,
		HavdalahDeg:      zmanim.Tzeit3SmallStars,
		NoModern:         true,
		ShabbatMevarchim: true,
	}

	return opts, tmpl, nil
}

func usage() string {
	prog := os.Args[0]
	return fmt.Sprintf("usage: %s template-path", prog)
}

func parseFile(tmpl *template.Template, fpath string) (*template.Template, error) {
	buf, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.Parse(string(buf))
	return tmpl, err
}

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

func HourOffset(z *zmanim.Zmanim, tz *time.Location, hours float64) time.Time {
	rise := z.Sunrise()
	seconds := rise.Unix() + int64(z.Hour()*hours)
	return time.Unix(seconds, 0).In(tz)
}

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

func DatePartsEqual(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func HDatePartsEqual(a, b hdate.HDate) bool {
	ay, am, ad := a.Day(), a.Month(), a.Year()
	by, bm, bd := b.Day(), b.Month(), b.Year()
	return ay == by && am == bm && ad == bd
}
