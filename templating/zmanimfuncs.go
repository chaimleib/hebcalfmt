package templating

import (
	"errors"
	"fmt"
	"time"

	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/molad"
	"github.com/hebcal/hebcal-go/zmanim"
)

// ZmanimFuncs builds a map of templating functions for zmanim
// using the given [hebcal.CalOptions].
func ZmanimFuncs(opts *hebcal.CalOptions) map[string]any {
	return map[string]any{
		// zmanim.Location
		"lookupCity":  LookupCity,
		"allCities":   zmanim.AllCities,
		"newLocation": zmanim.NewLocation,

		// zmanim.Zmanim
		"forDate":         ForDate(opts.Location, opts.UseElevation),
		"forLocationDate": ForLocationDate,

		// molad
		"molad": molad.New,
	}
}

// LookupCity is the same as [zmanim.LookupCity],
// except that we return an error if no match is found.
func LookupCity(city string) (*zmanim.Location, error) {
	l := zmanim.LookupCity(city)
	if l == nil {
		return nil, fmt.Errorf("unknown city %q", city)
	}
	return l, nil
}

// ForDate takes a zmanim.Location and a bool for use_elevation
// and returns a constructor for new zmanim.Zmanim objects,
// which takes time.Times in that Location.
// Unlike zmanim.New which can panic and returns a struct,
// this constructor returns a struct pointer and an error.
// Also unlike zmanim.New, it assigns the result's UseElevation field.
func ForDate(
	loc *zmanim.Location,
	useElevation bool,
) func(d time.Time) (*zmanim.Zmanim, error) {
	return func(d time.Time) (*zmanim.Zmanim, error) {
		if loc == nil {
			return nil, errors.New("provided location was nil")
		}

		_, err := time.LoadLocation(loc.TimeZoneId)
		if err != nil {
			return nil, err
		}

		z := zmanim.New(loc, d)
		z.UseElevation = useElevation
		return &z, nil
	}
}

// ForLocationDate creates a new zmanim.Zmanim object.
// Unlike zmanim.New which can panic and returns a struct,
// this constructor returns a struct pointer and an error.
// Also unlike zmanim.New, it assigns the result's UseElevation field.
func ForLocationDate(
	loc *zmanim.Location,
	useElevation bool,
	d time.Time,
) (*zmanim.Zmanim, error) {
	return ForDate(loc, useElevation)(d)
}
