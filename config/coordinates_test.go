package config_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/config"
)

func TestCoordinates_Validate(t *testing.T) {
	cases := []struct {
		Name  string
		Input *config.Coordinates
		Err   string
	}{
		{"nil", nil, ""},
		{"origin", &config.Coordinates{0, 0}, ""},
		{"quadrant I", &config.Coordinates{45, 90}, ""},
		{"quadrant II", &config.Coordinates{-45, 90}, ""},
		{"quadrant III", &config.Coordinates{-45, -90}, ""},
		{"quadrant IV", &config.Coordinates{45, -90}, ""},
		{
			"exceed max Lon",
			&config.Coordinates{0, 200},
			"invalid longitude: 200.000000",
		},
		{
			"subceed min Lon",
			&config.Coordinates{0, -200},
			"invalid longitude: -200.000000",
		},
		{
			"exceed max Lat",
			&config.Coordinates{100, 0},
			"invalid latitude: 100.000000",
		},
		{
			"subceed max Lat",
			&config.Coordinates{-100, 0},
			"invalid latitude: -100.000000",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := c.Input.Validate()
			if c.Err == "" {
				if got != nil {
					t.Errorf("got unexpected error: %v", got)
				}
				return
			}
			if c.Err != got.Error() {
				t.Errorf("errors do not match - want:\n%s\ngot:\n%v", c.Err, got)
			}
		})
	}
}
