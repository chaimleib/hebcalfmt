package config

import "fmt"

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

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
