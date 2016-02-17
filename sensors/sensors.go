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
