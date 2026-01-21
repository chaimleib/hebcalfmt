package templating

import "time"

var TimeConsts = map[string]any{
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
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
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
}
