// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"errors"
	"strings"
)

type Speed struct {
	metersPerSecond float64
}

func NewSpeedMetersPerSecond(value float64) Speed {
	return Speed{value}
}

func (s *Speed) MetersPerSecond() float64 {
	return s.metersPerSecond
}

func (s *Speed) KilometersPerHour() float64 {
	return s.metersPerSecond * 3.6
}

func (s *Speed) MilesPerHour() float64 {
	return s.metersPerSecond * 2.2369363
}

func (s *Speed) Knots() float64 {
	return s.metersPerSecond * 1.9438445
}

func (s *Speed) FeetPerSecond() float64 {
	return s.metersPerSecond * 3.2808399
}

func (s *Speed) Get(unit string) (float64, error) {
	switch strings.ToLower(unit) {
	case "m/s":
		return s.MetersPerSecond(), nil
	case "km/h":
		return s.KilometersPerHour(), nil
	case "mph":
		return s.MilesPerHour(), nil
	case "knots", "kts":
		return s.Knots(), nil
	case "ft/s":
		return s.FeetPerSecond(), nil
	}
	return 0, errors.New("Unknown unit")
}
