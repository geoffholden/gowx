// Copyright © 2017 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"testing"
	"testing/quick"
)

func TestSpeedMetersPerSecond(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewSpeedMetersPerSecond(x)
		return floatEquals(x, y.MetersPerSecond())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestSpeedKilometersPerHour(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewSpeedKilometersPerHour(x)
		return floatEquals(x, y.KilometersPerHour())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestSpeedGet(t *testing.T) {
	speed := NewSpeedMetersPerSecond(1)

	value, err := speed.Get("m/s")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 1) {
		t.Fatal("Value should be 1")
	}

	value, err = speed.Get("km/h")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 3.6) {
		t.Fatal("Value should be 3.6")
	}

	value, err = speed.Get("mph")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 2.2369363) {
		t.Fatal("Value should be 2.2369363")
	}

	value, err = speed.Get("M")
	if err == nil {
		t.Fatal("Invalid unit should give an error")
	}
}
