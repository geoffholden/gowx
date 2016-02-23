// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package sensors

import (
	"github.com/geoffholden/gowx/data"
	"github.com/spf13/viper"
	"math"
	"strconv"
	"strings"
	"time"
)

type BMP struct {
	data.SensorParser

	ossMode  int32
	avgCount int32
	cal      [11]int32

	c5, c6, mc, md float64
	x0, x1, x2     float64
	y0, y1, y2     float64
	p0, p1, p2     float64

	calibrated bool
}

func init() {
	var b BMP
	RegisterSensor("BMO", &b)
	RegisterSensor("BMV", &b)
	RegisterSensor("BM0", &b)
	RegisterSensor("BM1", &b)
	RegisterSensor("BM2", &b)
	RegisterSensor("BM3", &b)
	RegisterSensor("BM4", &b)
	RegisterSensor("BM5", &b)
	RegisterSensor("BM6", &b)
	RegisterSensor("BM7", &b)
	RegisterSensor("BM8", &b)
	RegisterSensor("BM9", &b)
	RegisterSensor("BMA", &b)
	RegisterSensor("BMX", &b)
}

func parseSignedShort(s string) int16 {
	val, err := strconv.ParseUint(s, 16, 16)
	if err != nil {
		panic(err)
	}

	var result int16
	result = int16(val)
	if val > 32767 {
		result = -1*^int16(val) - 1
	}

	return result
}

func (b *BMP) Parse(key string, input string) data.SensorData {
	switch key {
	case "BM0", "BM1", "BM2", "BM6", "BM7", "BM8", "BM9":
		val := parseSignedShort(input)
		b.cal[key[2]-'0'] = int32(val)
	case "BM3", "BM4", "BM5":
		val, err := strconv.ParseUint(input, 16, 16)
		if err != nil {
			panic(err)
		}
		b.cal[key[2]-'0'] = int32(val)
	case "BMA":
		b.cal[10] = int32(parseSignedShort(input))
		b.updateCal()
	case "BMV":
		val, err := strconv.ParseInt(input, 16, 16)
		if err != nil {
			panic(err)
		}
		b.avgCount = int32(val)
	case "BMO":
		val, err := strconv.ParseInt(input, 16, 16)
		if err != nil {
			panic(err)
		}
		b.ossMode = int32(val)
	case "BMX":
		if !b.calibrated {
			return data.SensorData{}
		}
		str := strings.Split(input, ",")
		temp, _ := strconv.ParseUint(str[0], 16, 32)
		pres, _ := strconv.ParseUint(str[1], 16, 32)

		t := uint32(temp) / uint32(b.avgCount)
		p := uint32(pres) / uint32(b.avgCount*16.0)

		alpha := b.c5 * (float64(t) - b.c6)
		bmpTemperature := alpha + b.mc/(alpha+b.md)

		s := bmpTemperature - 25.0
		x := (b.x2*s+b.x1)*s + b.x0
		y := (b.y2*s+b.y1)*s + b.y0
		z := (float64(p) - x) / y
		bmpPressure := (b.p2*z+b.p1)*z + b.p0

		elevationAdj := float64(viper.GetInt("elevation")) * 12.0 / 100.0 // 12 hPa/100m
		bmpPressure += elevationAdj

		var result data.SensorData
		result.TimeStamp = time.Now().UTC()
		result.ID = "BMP"
		result.Channel = 0
		result.Serial = "0"
		result.Data = make(map[string]float64)
		result.Data["Temperature"] = float64(bmpTemperature)
		result.Data["Pressure"] = float64(bmpPressure)
		return result
	}
	return data.SensorData{}
}

func (b *BMP) updateCal() {
	c3 := 160.0 * math.Pow(2, -15) * float64(b.cal[2])
	c4 := 0.001 * math.Pow(2, -15) * float64(b.cal[3])
	b1 := 160 * 160 * math.Pow(2, -30) * float64(b.cal[6])

	b.c5 = math.Pow(2, -15) / 160.0 * float64(b.cal[4])
	b.c6 = float64(b.cal[5])
	b.mc = math.Pow(2, 11) / (160 * 160) * float64(b.cal[9])
	b.md = float64(b.cal[10]) / 160.0

	b.x0 = float64(b.cal[0])
	b.x1 = 160.0 * math.Pow(2, -13) * float64(b.cal[1])
	b.x2 = 160 * 160 * math.Pow(2, -25) * float64(b.cal[7])

	b.y0 = c4 * math.Pow(2, 15)
	b.y1 = c4 * c3
	b.y2 = c4 * b1

	b.p0 = (3791.0 - 8.0) / 1600.0
	b.p1 = 1.0 - 7357.0*math.Pow(2, -20)
	b.p2 = 3038 * 100 * math.Pow(2, -36)

	b.calibrated = true
}
