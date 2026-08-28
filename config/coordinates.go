package config

import "fmt"

// Coordinates holds a latitude-longitude pair and an elevation in meters.
type Coordinates struct {
	// Lat is positive for North.
	Lat float64 `json:"lat"`

	// Lon is positive for East.
	Lon float64 `json:"lon"`

	// Elevation is in meters.
	Elevation int `json:"elevation"`
}

// Validate returns an error if the `Lat` or `Lon` field is out of bounds,
// the `Elevation` is negative,
// or if the [Coordinates] is nil.
func (c *Coordinates) Validate() error {
	if c == nil {
		return nil
	}
	if c.Lon < -180 || c.Lon > 180 {
		return fmt.Errorf("invalid longitude: %f", c.Lon)
	}
	if c.Lat < -90 || c.Lat > 90 {
		return fmt.Errorf("invalid latitude: %f", c.Lat)
	}
	if c.Elevation < 0 {
		return fmt.Errorf("negative elevation: %d", c.Elevation)
	}
	return nil
}
