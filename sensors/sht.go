// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package sensors

import (
	"github.com/geoffholden/gowx/data"
	"strconv"
	"strings"
	"time"
)

type SHT struct {
	data.SensorParser
}

func init() {
	var s SHT
	RegisterSensor("SHT", &s)
	RegisterSensor("DHT", &s)
}

func (d *SHT) Parse(key string, input string) data.SensorData {
	str := strings.Split(input, ",")
	temp, _ := strconv.ParseInt(str[0], 16, 16)
	hum, _ := strconv.ParseInt(str[1], 16, 16)

	var result data.SensorData
	result.TimeStamp = time.Now().UTC()
	result.ID = key
	result.Channel = 0
	result.Serial = "0"
	result.Data = make(map[string]float64)
	result.Data["Temperature"] = float64(temp) / 10.0
	result.Data["Humidity"] = float64(hum) / 10.0
	return result
}
