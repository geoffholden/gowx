package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/tarm/serial"
)

var sensors map[string]SensorParser

type SensorParser interface {
	Parse(key string, data string)
}

type Parser struct {
}

func RegisterSensor(key string, sensor SensorParser) {
	if nil == sensors {
		sensors = make(map[string]SensorParser)
	}
	sensors[key] = sensor
}

func (p *Parser) Loop(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.SplitN(scanner.Text(), ":", 2)
		if nil == sensors[line[0]] {
			fmt.Println(scanner.Text())
		} else {
			sensors[line[0]].Parse(line[0], line[1])
		}
	}
}

func (p *Parser) SerialLoop(port string) {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 57600}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
	p.Loop(s)
}
