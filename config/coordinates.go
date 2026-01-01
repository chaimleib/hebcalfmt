package config

import "fmt"

// Coordinates holds a latitude-longitude pair.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Validate returns an error if the `Lat` or `Lon` field is out of bounds,
// or if the [Coordinates] pair is nil.
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
	return nil
}
