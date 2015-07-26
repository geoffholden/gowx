package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type SHT struct {
	SensorParser
}

func init() {
	var s SHT
	RegisterSensor("SHT", &s)
	RegisterSensor("DHT", &s)
}

func (d *SHT) Parse(data string) {
	fmt.Print("SHT Sensor - ")
	str := strings.Split(data, ",")
	temp, _ := strconv.ParseInt(str[0], 16, 16)
	hum, _ := strconv.ParseInt(str[1], 16, 16)
	fmt.Printf("Temperature: %0.1fC ", float32(temp)/10)
	fmt.Printf("Humidity: %0.1f%%", float32(hum)/10)
	fmt.Println("")
}
