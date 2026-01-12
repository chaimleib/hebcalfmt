package test_test

import (
	"testing"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/yerushalmi"
	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/test"
)

func TestCheckCalOptions(t *testing.T) {
	cases := []struct {
		Name                string
		WantInput, GotInput *hebcal.CalOptions
		Logs                string
	}{
		{Name: "empties"},
		{
			Name:      "want nil, got non-nil",
			WantInput: nil,
			GotInput:  &hebcal.CalOptions{},
			Logs:      "expected nil, got: &hebcal.CalOptions{Location:(*zmanim.Location)(nil), Year:0, IsHebrewYear:false, NoJulian:false, Month:0, NumYears:0, Start:hdate.HDate{year:0, month:0, day:0, abs:0}, End:hdate.HDate{year:0, month:0, day:0, abs:0}, CandleLighting:false, CandleLightingMins:0, HavdalahMins:0, HavdalahDeg:0, Sedrot:false, IL:false, NoMinorFast:false, NoModern:false, NoRoshChodesh:false, ShabbatMevarchim:false, NoSpecialShabbat:false, NoHolidays:false, DafYomi:false, MishnaYomi:false, YerushalmiYomi:false, NachYomi:false, YerushalmiEdition:0, Omer:false, Molad:false, AddHebrewDates:false, AddHebrewDatesForEvents:false, Mask:0x0, YomKippurKatan:false, Hour24:false, SunriseSunset:false, DailyZmanim:false, Yahrzeits:[]hebcal.UserYahrzeit(nil), UserEvents:[]hebcal.UserEvent(nil), WeeklyAbbreviated:false, DailySedra:false}\n",
		},
		{
			Name:      "want non-nil, got nil",
			WantInput: &hebcal.CalOptions{},
			GotInput:  nil,
			Logs:      "got nil, want: &hebcal.CalOptions{Location:(*zmanim.Location)(nil), Year:0, IsHebrewYear:false, NoJulian:false, Month:0, NumYears:0, Start:hdate.HDate{year:0, month:0, day:0, abs:0}, End:hdate.HDate{year:0, month:0, day:0, abs:0}, CandleLighting:false, CandleLightingMins:0, HavdalahMins:0, HavdalahDeg:0, Sedrot:false, IL:false, NoMinorFast:false, NoModern:false, NoRoshChodesh:false, ShabbatMevarchim:false, NoSpecialShabbat:false, NoHolidays:false, DafYomi:false, MishnaYomi:false, YerushalmiYomi:false, NachYomi:false, YerushalmiEdition:0, Omer:false, Molad:false, AddHebrewDates:false, AddHebrewDatesForEvents:false, Mask:0x0, YomKippurKatan:false, Hour24:false, SunriseSunset:false, DailyZmanim:false, Yahrzeits:[]hebcal.UserYahrzeit(nil), UserEvents:[]hebcal.UserEvent(nil), WeeklyAbbreviated:false, DailySedra:false}\n",
		},
		{
			Name:      "mismatching Yahrzeits",
			WantInput: &hebcal.CalOptions{},
			GotInput: &hebcal.CalOptions{
				Yahrzeits: []hebcal.UserYahrzeit{
					{
						Date: time.Date(
							1929, time.May, 3, 0, 0, 0, 0, time.UTC),
						Name: "Joe Shmo",
					},
				},
			},
			Logs: `Yahrzeits's do not match - want:
[]
got:
[{1929-05-03 00:00:00 +0000 UTC Joe Shmo}]
`,
		},
		{
			Name: "mismatching UserEvents",
			WantInput: &hebcal.CalOptions{
				UserEvents: []hebcal.UserEvent{
					{
						Day:   10,
						Month: hdate.Tamuz,
						Desc:  "Jane Doe",
					},
				},
			},
			GotInput: &hebcal.CalOptions{},
			Logs: `UserEvents's do not match - want:
[{Tammuz 10 Jane Doe}]
got:
[]
`,
		},
		{
			Name: "full",
			WantInput: &hebcal.CalOptions{
				Location:     zmanim.LookupCity("Jerusalem"),
				Year:         5740,
				IsHebrewYear: true,
				NoJulian:     true,
				// Conflicts with IsHebrewYear,
				// but setting anyway to exercise all fields.
				Month:    time.July,
				NumYears: 1,
				Start:    hdate.New(5740, hdate.Tamuz, 1),
				End: hdate.New(
					5740,
					hdate.Tamuz,
					hdate.DaysInMonth(hdate.Tamuz, 5740),
				),
				CandleLighting:          true,
				CandleLightingMins:      30,
				HavdalahMins:            72,
				HavdalahDeg:             8.5,
				Sedrot:                  true,
				IL:                      true,
				NoMinorFast:             true,
				NoModern:                true,
				NoRoshChodesh:           true,
				ShabbatMevarchim:        true,
				NoSpecialShabbat:        true,
				NoHolidays:              true,
				DafYomi:                 true,
				MishnaYomi:              true,
				YerushalmiYomi:          true,
				YerushalmiEdition:       yerushalmi.Schottenstein,
				Omer:                    true,
				Molad:                   true,
				AddHebrewDates:          true,
				AddHebrewDatesForEvents: true,
				Mask:                    0xffff,
				YomKippurKatan:          true,
				Hour24:                  true,
				SunriseSunset:           true,
				DailyZmanim:             true,
				Yahrzeits: []hebcal.UserYahrzeit{
					{
						Date: hdate.New(5739, hdate.Tamuz, 15).Gregorian(),
						Name: "Joe Shmo",
					},
				},
				UserEvents: []hebcal.UserEvent{
					{
						Day:   14,
						Month: hdate.Tamuz,
						Desc:  "Birthday - Jane Doe",
					},
				},
				WeeklyAbbreviated: true,
				DailySedra:        true,
			},
			GotInput: &hebcal.CalOptions{
				Location:     zmanim.LookupCity("Jerusalem"),
				Year:         5740,
				IsHebrewYear: true,
				NoJulian:     true,
				// Conflicts with IsHebrewYear,
				// but setting anyway to exercise all fields.
				Month:    time.July,
				NumYears: 1,
				Start:    hdate.New(5740, hdate.Tamuz, 1),
				End: hdate.New(
					5740,
					hdate.Tamuz,
					hdate.DaysInMonth(hdate.Tamuz, 5740),
				),
				CandleLighting:          true,
				CandleLightingMins:      30,
				HavdalahMins:            72,
				HavdalahDeg:             8.5,
				Sedrot:                  true,
				IL:                      true,
				NoMinorFast:             true,
				NoModern:                true,
				NoRoshChodesh:           true,
				ShabbatMevarchim:        true,
				NoSpecialShabbat:        true,
				NoHolidays:              true,
				DafYomi:                 true,
				MishnaYomi:              true,
				YerushalmiYomi:          true,
				YerushalmiEdition:       yerushalmi.Schottenstein,
				Omer:                    true,
				Molad:                   true,
				AddHebrewDates:          true,
				AddHebrewDatesForEvents: true,
				Mask:                    0xffff,
				YomKippurKatan:          true,
				Hour24:                  true,
				SunriseSunset:           true,
				DailyZmanim:             true,
				Yahrzeits: []hebcal.UserYahrzeit{
					{
						Date: hdate.New(5739, hdate.Tamuz, 15).Gregorian(),
						Name: "Joe Shmo",
					},
				},
				UserEvents: []hebcal.UserEvent{
					{
						Day:   14,
						Month: hdate.Tamuz,
						Desc:  "Birthday - Jane Doe",
					},
				},
				WeeklyAbbreviated: true,
				DailySedra:        true,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.CheckCalOptions(mockT, c.WantInput, c.GotInput)
			wantFailed := c.Logs != ""
			if wantFailed != mockT.Failed() {
				t.Errorf("wantFailed is %v, but t.Failed() is %v",
					wantFailed, mockT.Failed())
			}
			if gotLogs := mockT.buf.String(); c.Logs != gotLogs {
				t.Errorf("logs did not match - want:\n%s\ngot:\n%s",
					c.Logs, gotLogs)
			}
		})
	}
}
