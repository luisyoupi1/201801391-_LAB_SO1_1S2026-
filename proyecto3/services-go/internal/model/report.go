package model

import (
	"fmt"
	"time"
)

var Countries = []string{"USA", "RUS", "CHN", "ESP", "GTM"}

type WarReport struct {
	ID              string    `json:"id"`
	Country         string    `json:"country"`
	WarplanesInAir  int32     `json:"warplanes_in_air"`
	WarshipsInWater int32     `json:"warships_in_water"`
	Timestamp       time.Time `json:"timestamp"`
}

func (r WarReport) Validate() error {
	validCountry := false
	for _, country := range Countries {
		if r.Country == country {
			validCountry = true
			break
		}
	}
	if !validCountry {
		return fmt.Errorf("invalid country %q", r.Country)
	}
	if r.WarplanesInAir < 0 || r.WarplanesInAir > 50 {
		return fmt.Errorf("warplanes_in_air outside 0..50")
	}
	if r.WarshipsInWater < 0 || r.WarshipsInWater > 30 {
		return fmt.Errorf("warships_in_water outside 0..30")
	}
	if r.ID == "" || r.Timestamp.IsZero() {
		return fmt.Errorf("id and timestamp are required")
	}
	return nil
}
