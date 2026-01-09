package test_test

import (
	"testing"
	"time"

	"github.com/hebcal/hdate"

	"github.com/chaimleib/hebcalfmt/daterange"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestCheckDateRangeSource(t *testing.T) {
	now := time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC)
	otherTime := time.Date(2022, 5, 6, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		Name                string
		WantInput, GotInput daterange.Source
		Failed              bool
		Logs                string
	}{
		{Name: "empties"},
		{
			Name:      "IsHebrewDate vs !IsHebrewDate",
			WantInput: daterange.Source{IsHebrewDate: true},
			GotInput:  daterange.Source{IsHebrewDate: false},
			Failed:    true,
			Logs:      "Source.IsHebrewDate's do not match - want: true, got: false\n",
		},
		{
			Name:      "Now vs same Now",
			WantInput: daterange.Source{Now: now},
			GotInput:  daterange.Source{Now: now},
		},
		{
			Name:      "Now vs other time",
			WantInput: daterange.Source{Now: now},
			GotInput:  daterange.Source{Now: otherTime},
			Failed:    true,
			Logs: `Source.Now's do not match - want:
2023-03-05 00:00:00 +0000 UTC
got:
2022-05-06 00:00:00 +0000 UTC
`,
		},
		{
			Name:      "FromTime vs nil",
			WantInput: daterange.Source{FromTime: &now},
			GotInput:  daterange.Source{FromTime: nil},
			Failed:    true,
			Logs: `Source.FromTime's nilness do not match - want:
2023-03-05 00:00:00 +0000 UTC
got:
<nil>
`,
		},
		{
			Name:      "FromTime vs same FromTime",
			WantInput: daterange.Source{FromTime: &now},
			GotInput:  daterange.Source{FromTime: &now},
		},
		{
			Name:      "FromTime vs other time",
			WantInput: daterange.Source{FromTime: &now},
			GotInput:  daterange.Source{FromTime: &otherTime},
			Failed:    true,
			Logs: `Source.FromTime's values do not match - want:
2023-03-05 00:00:00 +0000 UTC
got:
2022-05-06 00:00:00 +0000 UTC
`,
		},
		{
			Name:      "empty Args vs nil",
			WantInput: daterange.Source{Args: []string{}},
			GotInput:  daterange.Source{Args: nil},
			Failed:    true,
			Logs:      "Source.Args's nilness do not match - want==nil: false, got==nil: true\n",
		},
		{
			Name:      "1 Arg vs same 1 Arg",
			WantInput: daterange.Source{Args: []string{"2023"}},
			GotInput:  daterange.Source{Args: []string{"2023"}},
		},
		{
			Name:      "1 Arg vs different 1 Arg",
			WantInput: daterange.Source{Args: []string{"2024"}},
			GotInput:  daterange.Source{Args: []string{"2021"}},
			Failed:    true,
			Logs: `unexpected Source.Args[0] - want: "2024", got: "2021"
`,
		},
		{
			Name:      "1 Arg vs 2 Args",
			WantInput: daterange.Source{Args: []string{"2024"}},
			GotInput:  daterange.Source{Args: []string{"2024", "7"}},
			Failed:    true,
			Logs: `length of Source.Args does not match - want: 1, got: 2
unexpected extra Source.Args at index 1, skipping rest - got: "7"
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.CheckDateRangeSource(mockT, c.WantInput, c.GotInput)
			if c.Failed != mockT.Failed() {
				t.Errorf("c.Failed is %v, but t.Failed() is %v",
					c.Failed, mockT.Failed())
			}
			if gotLogs := mockT.buf.String(); c.Logs != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Logs, gotLogs)
			}
		})
	}
}

func TestCheckDateRange(t *testing.T) {
	cases := []struct {
		Name                string
		WantInput, GotInput daterange.DateRange
		Failed              bool
		Logs                string
	}{
		{Name: "empties"},
		{
			Name:      "rangetype same",
			WantInput: daterange.DateRange{RangeType: daterange.RangeTypeMonth},
			GotInput:  daterange.DateRange{RangeType: daterange.RangeTypeMonth},
		},
		{
			Name:      "rangetype different",
			WantInput: daterange.DateRange{RangeType: daterange.RangeTypeMonth},
			GotInput:  daterange.DateRange{RangeType: daterange.RangeTypeDay},
			Failed:    true,
			Logs:      "DateRange.RangeType's did not match - want: MONTH, got: DAY\n",
		},
		{
			Name:      "day same",
			WantInput: daterange.DateRange{Day: 1},
			GotInput:  daterange.DateRange{Day: 1},
		},
		{
			Name:      "day different",
			WantInput: daterange.DateRange{Day: 1},
			GotInput:  daterange.DateRange{Day: 20},
			Failed:    true,
			Logs:      "DateRange.Day's did not match - want: 1, got: 20\n",
		},
		{
			Name:      "gregMonth same",
			WantInput: daterange.DateRange{GregMonth: time.May},
			GotInput:  daterange.DateRange{GregMonth: time.May},
		},
		{
			Name:      "gregMonth different",
			WantInput: daterange.DateRange{GregMonth: time.May},
			GotInput:  daterange.DateRange{GregMonth: time.November},
			Failed:    true,
			Logs:      "DateRange.GregMonth's did not match - want: May, got: November\n",
		},
		{
			Name:      "hebMonth same",
			WantInput: daterange.DateRange{HebMonth: hdate.Iyyar},
			GotInput:  daterange.DateRange{HebMonth: hdate.Iyyar},
		},
		{
			Name:      "hebMonth different",
			WantInput: daterange.DateRange{HebMonth: hdate.Sivan},
			GotInput:  daterange.DateRange{HebMonth: hdate.Tevet},
			Failed:    true,
			Logs:      "DateRange.HebMonth's did not match - want: Sivan, got: Tevet\n",
		},
		{
			Name:      "year same",
			WantInput: daterange.DateRange{Year: 2020},
			GotInput:  daterange.DateRange{Year: 2020},
		},
		{
			Name:      "year different",
			WantInput: daterange.DateRange{Year: 2020},
			GotInput:  daterange.DateRange{Year: 1920},
			Failed:    true,
			Logs:      "DateRange.Year's did not match - want: 2020, got: 1920\n",
		},
		{
			Name:      "isHebrewDate same",
			WantInput: daterange.DateRange{IsHebrewDate: true},
			GotInput:  daterange.DateRange{IsHebrewDate: true},
		},
		{
			Name:      "isHebrewDate different",
			WantInput: daterange.DateRange{IsHebrewDate: true},
			GotInput:  daterange.DateRange{IsHebrewDate: false},
			Failed:    true,
			Logs:      "DateRange.IsHebrewDate's did not match - want: true, got: false\n",
		},
		{
			Name: "sources same",
			WantInput: daterange.DateRange{
				Source: daterange.Source{IsHebrewDate: true},
			},
			GotInput: daterange.DateRange{
				Source: daterange.Source{IsHebrewDate: true},
			},
		},
		{
			Name: "sources different",
			WantInput: daterange.DateRange{
				Source: daterange.Source{IsHebrewDate: true},
			},
			GotInput: daterange.DateRange{
				Source: daterange.Source{IsHebrewDate: false},
			},
			Failed: true,
			Logs:   "Source.IsHebrewDate's do not match - want: true, got: false\n",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.CheckDateRange(mockT, "DateRange", c.WantInput, c.GotInput)
			if c.Failed != mockT.Failed() {
				t.Errorf("c.Failed is %v, but t.Failed() is %v",
					c.Failed, mockT.Failed())
			}
			if gotLogs := mockT.buf.String(); c.Logs != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Logs, gotLogs)
			}
		})
	}
}
