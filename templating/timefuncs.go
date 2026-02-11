package templating

import (
	"errors"
	"time"
)

var TimeFuncs = map[string]any{
	// time.Duration
	"secondsDuration":   SecondsDuration,
	"timeParseDuration": time.ParseDuration,
	"timeSince":         time.Since,
	"timeUntil":         time.Until,
	"durationDiv":       DurationDiv,
	"durationMul":       DurationMul,

	// time.Location
	"timeFixedZone":    time.FixedZone,
	"timeLoadLocation": time.LoadLocation,

	// time.Time
	"timeNow":             time.Now,
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

// DurationDiv divides d by the divisor.
func DurationDiv(d time.Duration, divisor float64) (time.Duration, error) {
	if divisor == 0 {
		return 0, errors.New("divide by zero")
	}
	return time.Duration(float64(d) / divisor), nil
}

// DurationMul multiplies d by the given factor.
func DurationMul(d time.Duration, factor float64) time.Duration {
	return time.Duration(float64(d) * factor)
}
