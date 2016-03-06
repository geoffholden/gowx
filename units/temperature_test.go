// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"testing"
	"testing/quick"
)

func TestTemperatureCelsius(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewTemperatureCelsius(x)
		return floatEquals(x, y.Celsius())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestTemperatureFahrenheit(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewTemperatureFahrenheit(x)
		return floatEquals(x, y.Fahrenheit())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestTemperatureKelvin(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewTemperatureKelvin(x)
		return floatEquals(x, y.Kelvin())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestTemperatureGet(t *testing.T) {
	temp := NewTemperatureCelsius(0)

	value, err := temp.Get("C")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 0) {
		t.Fatal("Value should be 0")
	}

	value, err = temp.Get("F")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 32) {
		t.Fatal("Value should be 32")
	}

	value, err = temp.Get("K")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 273.15) {
		t.Fatal("Value should be 273.15")
	}

	value, err = temp.Get("M")
	if err == nil {
		t.Fatal("Invalid unit should give an error")
	}
}
