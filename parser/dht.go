package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type DHT struct {
	SensorParser
}

func init() {
	RegisterSensor("DHT", new(DHT))
}

func (d *DHT) Parse(data string) {
	fmt.Print("DHT Sensor - ")
	str := strings.Split(data, ",")
	temp, _ := strconv.ParseInt(str[0], 16, 16)
	hum, _ := strconv.ParseInt(str[1], 16, 16)
	fmt.Printf("Temperature: %0.1fC ", float32(temp)/10)
	fmt.Printf("Humidity: %0.1f%%", float32(hum)/10)
	fmt.Println("")
}
