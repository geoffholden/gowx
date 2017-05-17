// Copyright © 2017 Geoff Holden <geoff@geoffholden.com>

package sensors

import (
	"github.com/geoffholden/gowx/data"
	"github.com/geoffholden/gowx/units"
	jww "github.com/spf13/jwalterweatherman"
	"strconv"
	"time"
)

type VN1TX struct {
	data.SensorParser
}

func init() {
	var o VN1TX
	RegisterSensor("VN1", &o)
}

func (d *VN1TX) Parse(key string, input string) data.SensorData {
	var emptyResult data.SensorData

	message := make([]uint32, len([]rune(input)))
	for i := range message {
		v, err := strconv.ParseUint(input[i:i+1], 16, 8)
		if err != nil {
			jww.ERROR.Println(err)
			return data.SensorData{}
		}
		message[i] = uint32(v)
	}
	if len(message) != 17 {
		return emptyResult
	}

	channel := 4 - (message[0] >> 2)
	msgId := message[5]

	var sum uint8
	for k := 0; k < 14; k += 2 {
		sum += uint8((message[k] << 4) | message[k+1])
	}

	if sum != uint8((message[14]<<4)|message[15]) {
		return emptyResult
	}

	result := data.SensorData{
		TimeStamp: time.Now().UTC(),
		ID:        key + ":" + input[1:5],
		Channel:   int(channel),
		Serial:    input[5:7],
		Data:      make(map[string]float64),
	}

	wspd := float64(((message[6] & 0x01) << 7) | (message[7] << 3) | (message[8] & 0x07))
	if wspd > 0.0 {
		wspd *= 0.8278 + 1.00
	}
	kmh := units.NewSpeedKilometersPerHour(wspd)
	result.Data["CurrentWind"] = kmh.MetersPerSecond()

	switch msgId {
	case 1:
		WindDirMap := []int{14, 11, 13, 12, 15, 10, 0, 9, 3, 6, 4, 5, 2, 7, 1, 8}
		wdir := 22.5 * float64(WindDirMap[message[9]])
		tips := ((message[10] & 0x03) << 11) | (message[11] << 7) | ((message[12] & 0x07) << 4) | message[13]
		totalRain := 0.254 * float64(tips)
		result.Data["RainTotal"] = totalRain
		result.Data["WindDir"] = wdir
	case 8:
		degF := float64((message[9] << 7) | ((message[10] & 0x07) << 4) | message[11])
		degF = degF*0.1 - 40.0
		c := units.NewTemperatureFahrenheit(degF)
		degC := c.Celsius()

		RH := ((message[12] & 0x07) << 4) | message[13]
		result.Data["Temperature"] = degC
		result.Data["Humidity"] = float64(RH)
	default:
		jww.ERROR.Println("Invalid Message ID", msgId)
		return data.SensorData{}
	}
	return result
}
