package templating

import (
	"github.com/hebcal/hdate"

	"github.com/chaimleib/hebcalfmt/xhdate"
)

var HDateFuncs = map[string]any{
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
}
