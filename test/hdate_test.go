package test_test

import (
	"testing"

	"github.com/hebcal/hdate"

	"github.com/chaimleib/hebcalfmt/test"
)

func TestCheckHDate(t *testing.T) {
	var zero hdate.HDate
	rh := hdate.New(5775, hdate.Tishrei, 1)
	rhGreg := rh.Gregorian()
	// rhConverted has cached the RD number,
	// so it is no longer struct-equal with rh
	rhConverted := hdate.FromTime(rhGreg)
	if rh == rhConverted {
		t.Errorf(
			"rh unexpectedly struct-equals rhConverted. Does it still cache the RD?\nrhConverted: %#v",
			rhConverted,
		)
	}
	chanukah := hdate.New(5775, hdate.Kislev, 25)
	_ = chanukah

	cases := []struct {
		Name                string
		WantInput, GotInput hdate.HDate
		Failed              bool
		Logs                string
	}{
		{Name: "empties"},
		{
			Name:      "simple vs zero",
			WantInput: rh,
			GotInput:  zero,
			Failed:    true,
			Logs: `HDate did not match - want:
1 Tishrei 5775
got:
0 %!HMonth(0) 0
`,
		},
		{
			Name:      "zero vs simple",
			WantInput: zero,
			GotInput:  rh,
			Failed:    true,
			Logs: `HDate did not match - want:
0 %!HMonth(0) 0
got:
1 Tishrei 5775
`,
		},
		{
			Name:      "simple",
			WantInput: rh,
			GotInput:  rh,
		},
		{
			Name:      "simple vs converted",
			WantInput: rh,
			GotInput:  rhConverted,
		},
		{
			Name:      "converted vs simple",
			WantInput: rhConverted,
			GotInput:  rh,
		},
		{
			Name:      "RH vs Chanukah",
			WantInput: rh,
			GotInput:  chanukah,
			Failed:    true,
			Logs: `HDate did not match - want:
1 Tishrei 5775
got:
25 Kislev 5775
`,
		},
		{
			Name:      "Chanukah vs RH",
			WantInput: chanukah,
			GotInput:  rh,
			Failed:    true,
			Logs: `HDate did not match - want:
25 Kislev 5775
got:
1 Tishrei 5775
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.CheckHDate(mockT, "HDate", c.WantInput, c.GotInput)

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
