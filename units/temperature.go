// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"errors"
	"strings"
)

type Temperature struct {
	celsius float64
}

func NewTemperatureCelsius(value float64) Temperature {
	return Temperature{value}
}

func NewTemperatureKelvin(value float64) Temperature {
	return Temperature{value - 273.15}
}

func NewTemperatureFahrenheit(value float64) Temperature {
	return Temperature{(value - 32) / 1.8}
}

func (t *Temperature) Celsius() float64 {
	return t.celsius
}

func (t *Temperature) Fahrenheit() float64 {
	return t.celsius*1.8 + 32
}

func (t *Temperature) Kelvin() float64 {
	return t.celsius + 273.15
}

func (t *Temperature) Get(unit string) (float64, error) {
	switch strings.ToLower(unit) {
	case "c", "celsius":
		return t.Celsius(), nil
	case "f", "fahrenheit":
		return t.Fahrenheit(), nil
	case "k", "kelvin":
		return t.Kelvin(), nil
	}
	return 0, errors.New("Unknown unit")
}
