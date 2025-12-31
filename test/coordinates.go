package test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/config"
)

func CheckCoordinates(t *testing.T, want, got *config.Coordinates) {
	t.Helper()
	if want == got {
		return
	}

	if want.Lat != got.Lat {
		t.Errorf("latitudes do not match - want: %v, got: %v", want.Lat, got.Lat)
	}

	if want.Lon != got.Lon {
		t.Errorf("longitudes do not match - want: %v, got: %v", want.Lon, got.Lon)
	}
}
