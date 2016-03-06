// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"testing"
	"testing/quick"
)

func TestDistanceMillimeters(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewDistanceMillimeters(x)
		return floatEquals(x, y.Millimeters())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestDistanceMeters(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewDistanceMeters(x)
		return floatEquals(x, y.Meters())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestDistanceKilometers(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewDistanceKilometers(x)
		return floatEquals(x, y.Kilometers())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestDistanceMiles(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewDistanceMiles(x)
		return floatEquals(x, y.Miles())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestDistanceInches(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewDistanceInches(x)
		return floatEquals(x, y.Inches())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestDistanceFeet(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewDistanceFeet(x)
		return floatEquals(x, y.Feet())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestDistanceGet(t *testing.T) {
	dist := NewDistanceMeters(1)

	value, err := dist.Get("m")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 1) {
		t.Fatal("Value should be 0")
	}

	value, err = dist.Get("mm")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 1000) {
		t.Fatal("Value should be 1000")
	}

	value, err = dist.Get("cm")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 100) {
		t.Fatal("Value should be 100")
	}

	value, err = dist.Get("km")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 0.001) {
		t.Fatal("Value should be 0.001")
	}

	dist = NewDistanceMiles(1)
	value, err = dist.Get("mi")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 1) {
		t.Fatal("Value should be 1")
	}

	value, err = dist.Get("in")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 63360) {
		t.Fatal("Value should be 63360")
	}

	value, err = dist.Get("ft")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 5280) {
		t.Fatal("Value should be 5280")
	}

	value, err = dist.Get("C")
	if err == nil {
		t.Fatal("Invalid unit should give an error")
	}
}
