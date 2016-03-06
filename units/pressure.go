// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"errors"
	"strings"
)

type Pressure struct {
	hectopascal float64
}

func NewPressureHectopascal(value float64) Pressure {
	return Pressure{value}
}

func NewPressureHpa(value float64) Pressure {
	return NewPressureHectopascal(value)
}

func NewPressureKilopascal(value float64) Pressure {
	return Pressure{value * 10.0}
}

func NewPressurePascal(value float64) Pressure {
	return Pressure{value / 100.0}
}

func (p *Pressure) Pascal() float64 {
	return p.hectopascal * 100.0
}

func (p *Pressure) Hectopascal() float64 {
	return p.hectopascal
}

func (p *Pressure) Kilopascal() float64 {
	return p.hectopascal / 10.0
}

func (p *Pressure) Millibar() float64 {
	return p.Hectopascal()
}

func (p *Pressure) Bar() float64 {
	return p.Millibar() / 1000.0
}

func (p *Pressure) Atmosphere() float64 {
	return p.Pascal() / 101325.0
}

func (p *Pressure) MillimeterMercury() float64 {
	return p.Pascal() / 133.322387415
}

func (p *Pressure) InchMercury() float64 {
	return p.Pascal() / 3386.389
}

func (p *Pressure) PoundSquareInch() float64 {
	return p.Pascal() / 6894.757
}

func (p *Pressure) Get(unit string) (float64, error) {
	switch strings.ToLower(unit) {
	case "pa", "pascal":
		return p.Pascal(), nil
	case "hpa", "hectopascal":
		return p.Hectopascal(), nil
	case "kpa", "kilopascal":
		return p.Kilopascal(), nil
	case "bar":
		return p.Bar(), nil
	case "mbar", "millibar":
		return p.Millibar(), nil
	case "atm", "atmosphere":
		return p.Atmosphere(), nil
	case "mmhg":
		return p.MillimeterMercury(), nil
	case "inhg":
		return p.InchMercury(), nil
	case "psi":
		return p.PoundSquareInch(), nil
	}
	return 0, errors.New("Unknown unit")
}
