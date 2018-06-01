// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"errors"
	"strings"
)

type Distance struct {
	meters float64
}

func NewDistanceMeters(value float64) Distance {
	return Distance{value}
}

func NewDistanceMillimeters(value float64) Distance {
	return Distance{value / 1000.0}
}

func NewDistanceKilometers(value float64) Distance {
	return Distance{value * 1000.0}
}

func NewDistanceMiles(value float64) Distance {
	return Distance{value * 1609.3440}
}

func NewDistanceInches(value float64) Distance {
	return Distance{value * 0.0254}
}

func NewDistanceFeet(value float64) Distance {
	return Distance{value * 0.3048}
}

func NewDistanceNauticalMiles(value float64) Distance {
	return Distance{value * 1852.0}
}

func (d *Distance) Meters() float64 {
	return d.meters
}

func (d *Distance) Kilometers() float64 {
	return d.meters / 1000.0
}

func (d *Distance) Miles() float64 {
	return d.meters / 1609.3440
}

func (d *Distance) Millimeters() float64 {
	return d.meters * 1000.0
}

func (d *Distance) Centimeters() float64 {
	return d.meters * 100.0
}

func (d *Distance) Inches() float64 {
	return d.meters / 0.0254
}

func (d *Distance) Feet() float64 {
	return d.meters / 0.3048
}

func (d *Distance) NauticalMiles() float64 {
	return d.meters / 1852.0
}

func (d *Distance) Get(unit string) (float64, error) {
	switch strings.ToLower(unit) {
	case "m":
		return d.Meters(), nil
	case "km":
		return d.Kilometers(), nil
	case "mm":
		return d.Millimeters(), nil
	case "cm":
		return d.Centimeters(), nil
	case "mi", "mile", "miles":
		return d.Miles(), nil
	case "in", "inch", "inches":
		return d.Inches(), nil
	case "ft", "feet":
		return d.Feet(), nil
	case "nm":
		return d.NauticalMiles(), nil
	}
	return 0, errors.New("Unknown unit")
}
