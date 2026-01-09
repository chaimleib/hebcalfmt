package test

import (
	"github.com/chaimleib/hebcalfmt/config"
)

func CheckCoordinates(
	t Test,
	name string,
	want, got config.Coordinates,
) {
	t.Helper()

	if want.Lat != got.Lat {
		t.Errorf("%s.Lat's did not match - want: %v, got: %v",
			name, want.Lat, got.Lat)
	}

	if want.Lon != got.Lon {
		t.Errorf("%s.Lon's did not match - want: %v, got: %v",
			name, want.Lon, got.Lon)
	}
}
