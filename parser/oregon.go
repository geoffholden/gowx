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
	var emptyResult gowx.SensorData
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

	if key == "OS3" {
		// checksum validation
		if len(data) < 2 {
			return emptyResult
		}
		sum := int8(0)
		for _, b := range data[0 : len(data)-2] {
			val, _ := strconv.ParseInt(string(b), 16, 8)
			sum += int8(val)
		}

		provided := make([]byte, 2)
		provided[0] = data[len(data)-1]
		provided[1] = data[len(data)-2]

		x, _ := strconv.ParseInt(string(provided), 16, 8)
		if int8(x) != sum {
			return emptyResult
		}
	}

	switch data[0:4] {
	case "1D20", "F824", "F8B4":
		if len(data) != 17 {
			return emptyResult
		}
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
		if len(data) != 14 {
			return emptyResult
		}
		temperature := float64(data[10]-'0') * 10.0
		temperature += float64(data[9]-'0') * 1.0
		temperature += float64(data[8]-'0') * 0.1
		if data[11] != '0' {
			temperature *= -1
		}

		result.Data["Temperature"] = temperature
	case "EC70":
		if len(data) != 14 {
			return emptyResult
		}
		uv := (data[9] - '0') * 10
		uv += (data[8] - '0') * 1

		result.Data["UV"] = float64(uv)
	case "D874":
		if len(data) != 15 {
			return emptyResult
		}
		uv := (data[12] - '0') * 10
		uv += (data[11] - '0') * 1

		result.Data["UV"] = float64(uv)
	case "1984", "1994":
		if len(data) != 19 {
			return emptyResult
		}
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
		if len(data) != 20 {
			return emptyResult
		}
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
		if len(data) != 18 {
			return emptyResult
		}
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
		if len(data) != 20 {
			return emptyResult
		}
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
