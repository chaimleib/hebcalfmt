package test_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/config"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestCheckCoordinates(t *testing.T) {
	cases := []struct {
		Name                string
		WantInput, GotInput config.Coordinates
		Failed              bool
		Logs                string
	}{
		{Name: "empties"},
		{
			Name:      "same coords",
			WantInput: config.Coordinates{Lat: 30, Lon: 40},
			GotInput:  config.Coordinates{Lat: 30, Lon: 40},
		},
		{
			Name:      "different lat",
			WantInput: config.Coordinates{Lat: -30, Lon: 40},
			GotInput:  config.Coordinates{Lat: 30, Lon: 40},
			Failed:    true,
			Logs:      "Geo.Lat's did not match - want: -30, got: 30\n",
		},
		{
			Name:      "different lon",
			WantInput: config.Coordinates{Lat: 30, Lon: -40},
			GotInput:  config.Coordinates{Lat: 30, Lon: 40},
			Failed:    true,
			Logs:      "Geo.Lon's did not match - want: -40, got: 40\n",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			test.CheckCoordinates(mockT, "Geo", c.WantInput, c.GotInput)

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
