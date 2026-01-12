package templating

import (
	"errors"
	"fmt"
	"time"

	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/zmanim"
)

func ZmanimFuncs(opts *hebcal.CalOptions) map[string]any {
	return map[string]any{
		// zmanim.Location
		"lookupCity":  LookupCity,
		"allCities":   zmanim.AllCities,
		"newLocation": zmanim.NewLocation,

		// zmanim.Zmanim
		"forDate":         ForDate(opts.Location),
		"forLocationDate": ForLocationDate,
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

// ForDate takes a zmanim.Location
// and returns a constructor for new zmanim.Zmanim objects
// with different dates in that Location.
// Unlike zmanim.New which can panic and returns a struct,
// this constructor returns a struct pointer and an error.
func ForDate(loc *zmanim.Location) func(d time.Time) (*zmanim.Zmanim, error) {
	return func(d time.Time) (*zmanim.Zmanim, error) {
		if loc == nil {
			return nil, errors.New("provided location was nil")
		}

		year, month, day := d.Date()
		tz, err := time.LoadLocation(loc.TimeZoneId)
		if err != nil {
			return nil, err
		}

		return &zmanim.Zmanim{
			Location: loc,
			Year:     year,
			Month:    month,
			Day:      day,
			TimeZone: tz,
		}, nil
	}
}

// ForLocationDate creates a new zmanim.Zmanim object.
// Unlike zmanim.New which can panic and returns a struct,
// this constructor returns a struct pointer and an error.
func ForLocationDate(
	loc *zmanim.Location,
	d time.Time,
) (*zmanim.Zmanim, error) {
	return ForDate(loc)(d)
}
