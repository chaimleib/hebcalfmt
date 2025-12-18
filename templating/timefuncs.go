package templating

import "time"

var TimeFuncs = map[string]any{
	// time.Duration
	"secondsDuration":   SecondsDuration,
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
}

// SecondsDuration converts float64 seconds into a [time.Duration].
func SecondsDuration(secs float64) time.Duration {
	return time.Duration(secs * 1_000_000_000)
}

// DatePartsEqual compares the year, month, and day of two [time.Time]s
// and returns true if those parts all match.
func DatePartsEqual(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
