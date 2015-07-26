package parser

import "fmt"
import "strconv"

type Oregon struct {
	SensorParser
}

func init() {
	RegisterSensor("OS3", new(Oregon))
}

func (d *Oregon) Parse(data string) {
	fmt.Println("Sensor ID", data[0:4], "Channel", data[4:5], "Rolling Code", data[5:7], "Flags", data[7:8])
	switch data[0:4] {
	case "1984":
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

		total := float32(data[17]-'0') * 100.0
		total += float32(data[16]-'0') * 10.0
		total += float32(data[15]-'0') * 1.0
		total += float32(data[14]-'0') * 0.100
		total += float32(data[13]-'0') * 0.010
		total += float32(data[12]-'0') * 0.001

		fmt.Println("Rain Rate", rate, "Total Rain", total)
	case "F824":
		temperature := float32(data[10]-'0') * 10.0
		temperature += float32(data[9]-'0') * 1.0
		temperature += float32(data[8]-'0') * 0.1
		if data[11] != '0' {
			temperature *= -1
		}

		humidity := float32(data[13]-'0') * 10.0
		humidity += float32(data[12]-'0') * 1.0
		fmt.Println("Temperature", temperature, "Humidity", humidity)
	default:
		fmt.Println("Oregon Sensor", data[0:4])
	}
}
