package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/geoffholden/gowx/gowx"
	"github.com/tarm/serial"
)

type Parser struct {
	config gowx.Config
}

var sensors map[string]gowx.SensorParser

func RegisterSensor(key string, sensor gowx.SensorParser) {
	if nil == sensors {
		sensors = make(map[string]gowx.SensorParser)
	}
	sensors[key] = sensor
}

func (p *Parser) loop(reader io.Reader, client *MQTT.Client) {
	channel := make(chan gowx.SensorData)
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			line := strings.SplitN(scanner.Text(), ":", 2)
			if nil == sensors[line[0]] {
				fmt.Println(scanner.Text())
			} else {
				d := sensors[line[0]].Parse(line[0], line[1])
				channel <- d
			}
		}
		close(channel)
	}()

	for data := range channel {
		topic := "/gowx/sample"
		buf := new(bytes.Buffer)
		encoder := json.NewEncoder(buf)
		encoder.Encode(data)
		payload := buf.Bytes()
		if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			fmt.Println("Failed to send message.")
			panic(token.Error())
		}
		fmt.Printf("Publishing %s -> %s\n", topic, buf.Bytes())
	}
}

func (p *Parser) serialLoop(client *MQTT.Client) {
	c := &serial.Config{
		Name: p.config.Parser.SerialPort,
		Baud: p.config.Parser.SerialBaud,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
	defer s.Close()
	s.Flush()
	p.loop(s, client)
}

func main() {
	flag.Parse()
	var p Parser
	p.config = gowx.NewConfig().LoadFile()

	opts := MQTT.NewClientOptions().AddBroker(p.config.Global.MQTTServer).SetClientID("parser").SetCleanSession(true)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(0)

	fi, err := os.Stat(p.config.Parser.SerialPort)
	if err != nil {
		panic(err)
	}
	if fi.Mode()&os.ModeType != 0 {
		p.serialLoop(client)
	} else {
		file, err := os.Open(p.config.Parser.SerialPort)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		p.loop(file, client)
	}
}
