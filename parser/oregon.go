package parser

import "fmt"
import "strconv"

type Oregon struct {
	SensorParser
}

func init() {
	var o Oregon
	RegisterSensor("OS3", &o)
	RegisterSensor("OS2", &o)
}

func (d *Oregon) Parse(data string) {
	fmt.Println("Sensor ID", data[0:4], "Channel", data[4:5], "Rolling Code", data[5:7], "Flags", data[7:8])
	switch data[0:4] {
	case "1D20", "F824", "F8B4":
		temperature := float32(data[10]-'0') * 10.0
		temperature += float32(data[9]-'0') * 1.0
		temperature += float32(data[8]-'0') * 0.1
		if data[11] != '0' {
			temperature *= -1
		}

		humidity := float32(data[13]-'0') * 10.0
		humidity += float32(data[12]-'0') * 1.0
		fmt.Println("Temperature", temperature, "Humidity", humidity)
	case "EC40", "C844":
		temperature := float32(data[10]-'0') * 10.0
		temperature += float32(data[9]-'0') * 1.0
		temperature += float32(data[8]-'0') * 0.1
		if data[11] != '0' {
			temperature *= -1
		}

		fmt.Println("Temperature", temperature)
	case "EC70":
		uv := (data[9] - '0') * 10
		uv += (data[8] - '0') * 1

		fmt.Println("UV", uv)
	case "D874":
		uv := (data[12] - '0') * 10
		uv += (data[11] - '0') * 1

		fmt.Println("UV", uv)
	case "1984", "1994":
		dir, _ := strconv.ParseInt(data[8:9], 16, 8)
		direction := float32(dir) * 22.5

		current := float32(data[13]-'0') * 10.0
		current += float32(data[12]-'0') * 1.0
		current += float32(data[11]-'0') * 0.1
		average := float32(data[16]-'0') * 10.0
		average += float32(data[15]-'0') * 1.0
		average += float32(data[14]-'0') * 0.1

		fmt.Println("Direction", direction, "Current", current, "Average", average)
	case "2914":
		rate := float32(data[11]-'0') * 10.0
		rate += float32(data[10]-'0') * 1.0
		rate += float32(data[9]-'0') * 0.10
		rate += float32(data[8]-'0') * 0.01
		rate *= 25.4

		total := float32(data[17]-'0') * 100.0
		total += float32(data[16]-'0') * 10.0
		total += float32(data[15]-'0') * 1.0
		total += float32(data[14]-'0') * 0.100
		total += float32(data[13]-'0') * 0.010
		total += float32(data[12]-'0') * 0.001
		total *= 25.4

		fmt.Println("Rain Rate", rate, "Total Rain", total)
	case "2D10":
		rate := float32(data[10]-'0') * 10.0
		rate += float32(data[9]-'0') * 1.0
		rate += float32(data[8]-'0') * 0.1

		total := float32(data[15]-'0') * 1000.0
		total += float32(data[14]-'0') * 100.0
		total += float32(data[13]-'0') * 10.0
		total += float32(data[12]-'0') * 1.0
		total += float32(data[11]-'0') * 0.1

		fmt.Println("Rain Rate", rate, "Total Rain", total)
	case "5D60":
		temperature := float32(data[10]-'0') * 10.0
		temperature += float32(data[9]-'0') * 1.0
		temperature += float32(data[8]-'0') * 0.1
		if data[11] != '0' {
			temperature *= -1
		}

		humidity := float32(data[13]-'0') * 10.0
		humidity += float32(data[12]-'0') * 1.0

		var comfort string
		switch data[14] {
		case '0':
			comfort = "Normal"
		case '4':
			comfort = "Comfortable"
		case '8':
			comfort = "Dry"
		case 'C':
			comfort = "Wet"
		}

		pressure := 856
		pressure += int((data[16] - '0') * 10)
		pressure += int((data[15] - '0') * 1)

		var forecast string
		switch data[17] {
		case '2':
			forecast = "Cloudy"
		case '3':
			forecast = "Rainy"
		case '6':
			forecast = "Partly Cloudy"
		case 'C':
			forecast = "Sunny"
		}

		fmt.Println("Temperature", temperature, "Humidity", humidity, "Comfort", comfort, "Pressure", pressure, "Forecast", forecast)
	default:
		fmt.Println("Oregon Sensor", data[0:4])
	}
}
