// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"testing"
	"testing/quick"
)

func TestPressureHectopascal(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewPressureHectopascal(x)
		return floatEquals(x, y.Hectopascal())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestPressureKilopascal(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewPressureKilopascal(x)
		return floatEquals(x, y.Kilopascal())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestPressurePascal(t *testing.T) {
	if err := quick.Check(func(x float64) bool {
		y := NewPressurePascal(x)
		return floatEquals(x, y.Pascal())
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestPressureGet(t *testing.T) {
	temp := NewPressureKilopascal(101.325)

	value, err := temp.Get("kpa")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 101.325) {
		t.Fatal("Value should be 101.325")
	}

	value, err = temp.Get("atm")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 1) {
		t.Fatal("Value should be 1")
	}

	value, err = temp.Get("inhg")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 29.921373) {
		t.Fatal("Value should be 29.921373", value)
	}

	value, err = temp.Get("mmhg")
	if err != nil {
		t.Fatal(err)
	}
	if !floatEquals(value, 760) {
		t.Fatal("Value should be 760", value)
	}

	value, err = temp.Get("M")
	if err == nil {
		t.Fatal("Invalid unit should give an error")
	}
}
