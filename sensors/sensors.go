// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package sensors

import (
	"github.com/geoffholden/gowx/data"
)

var Sensors map[string]data.SensorParser

func RegisterSensor(key string, sensor data.SensorParser) {
	if nil == Sensors {
		Sensors = make(map[string]data.SensorParser)
	}
	Sensors[key] = sensor
}
