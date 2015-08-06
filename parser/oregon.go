package main

import "github.com/geoffholden/gowx/gowx"
import "strconv"
import "time"

type Oregon struct {
	gowx.SensorParser
}

func init() {
	var o Oregon
	RegisterSensor("OS3", &o)
	RegisterSensor("OS2", &o)
}

func (d *Oregon) Parse(key string, data string, config *gowx.Config) gowx.SensorData {
	//fmt.Println("Sensor ID", data[0:4], "Channel", data[4:5], "Rolling Code", data[5:7], "Flags", data[7:8])
	channel, err := strconv.ParseUint(data[4:5], 16, 8)
	if err != nil {
		panic(err)
	}
	result := gowx.SensorData{
		TimeStamp: time.Now().UTC(),
		ID:        key + ":" + data[0:4],
		Channel:   int(channel),
		Serial:    data[5:7],
		Data:      make(map[string]float64),
	}
	switch data[0:4] {
	case "1D20", "F824", "F8B4":
		temperature := float64(data[10]-'0') * 10.0
		temperature += float64(data[9]-'0') * 1.0
		temperature += float64(data[8]-'0') * 0.1
		if data[11] != '0' {
			temperature *= -1
		}

		humidity := float64(data[13]-'0') * 10.0
		humidity += float64(data[12]-'0') * 1.0
		result.Data["Temperature"] = temperature
		result.Data["Humidity"] = humidity
	case "EC40", "C844":
		temperature := float64(data[10]-'0') * 10.0
		temperature += float64(data[9]-'0') * 1.0
		temperature += float64(data[8]-'0') * 0.1
		if data[11] != '0' {
			temperature *= -1
		}

		result.Data["Temperature"] = temperature
	case "EC70":
		uv := (data[9] - '0') * 10
		uv += (data[8] - '0') * 1

		result.Data["UV"] = float64(uv)
	case "D874":
		uv := (data[12] - '0') * 10
		uv += (data[11] - '0') * 1

		result.Data["UV"] = float64(uv)
	case "1984", "1994":
		dir, _ := strconv.ParseInt(data[8:9], 16, 8)
		direction := float64(dir) * 22.5

		current := float64(data[13]-'0') * 10.0
		current += float64(data[12]-'0') * 1.0
		current += float64(data[11]-'0') * 0.1
		average := float64(data[16]-'0') * 10.0
		average += float64(data[15]-'0') * 1.0
		average += float64(data[14]-'0') * 0.1

		result.Data["WindDir"] = direction
		result.Data["CurrentWind"] = current
		result.Data["AverageWind"] = average
	case "2914":
		rate := float64(data[11]-'0') * 10.0
		rate += float64(data[10]-'0') * 1.0
		rate += float64(data[9]-'0') * 0.10
		rate += float64(data[8]-'0') * 0.01
		rate *= 25.4

		total := float64(data[17]-'0') * 100.0
		total += float64(data[16]-'0') * 10.0
		total += float64(data[15]-'0') * 1.0
		total += float64(data[14]-'0') * 0.100
		total += float64(data[13]-'0') * 0.010
		total += float64(data[12]-'0') * 0.001
		total *= 25.4

		result.Data["RainRate"] = rate
		result.Data["RainTotal"] = total
	case "2D10":
		rate := float64(data[10]-'0') * 10.0
		rate += float64(data[9]-'0') * 1.0
		rate += float64(data[8]-'0') * 0.1

		total := float64(data[15]-'0') * 1000.0
		total += float64(data[14]-'0') * 100.0
		total += float64(data[13]-'0') * 10.0
		total += float64(data[12]-'0') * 1.0
		total += float64(data[11]-'0') * 0.1

		result.Data["RainRate"] = rate
		result.Data["RainTotal"] = total
	case "5D60":
		temperature := float64(data[10]-'0') * 10.0
		temperature += float64(data[9]-'0') * 1.0
		temperature += float64(data[8]-'0') * 0.1
		if data[11] != '0' {
			temperature *= -1
		}

		humidity := float64(data[13]-'0') * 10.0
		humidity += float64(data[12]-'0') * 1.0

		pressure := 856.0
		pressure += float64((data[16] - '0') * 10)
		pressure += float64((data[15] - '0') * 1)

		result.Data["Temperature"] = temperature
		result.Data["Humidity"] = humidity
		result.Data["Pressure"] = pressure
	default:
	}
	return result
}
