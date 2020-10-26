// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/geoffholden/gowx/data"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"

	"go.bug.st/serial"

	"github.com/geoffholden/gowx/sensors"
)

// parserCmd represents the parser command
var parserCmd = &cobra.Command{
	Use:   "parser",
	Short: "Parse serial data",
	Long:  `Parses the serial data coming from the WSDL sheild and sends it to the MQTT broker.`,
	Run:   parser,
}

func parserInit() {
	if !parserCmd.Flags().HasFlags() {
		parserCmd.Flags().String("port", "", "Serial port to connect to")
		parserCmd.Flags().Int("baud", 9600, "Serial baud rate")
		parserCmd.Flags().Int("elevation", 0, "Elevation above sea level")
	}
}

func init() {
	RootCmd.AddCommand(parserCmd)
	parserInit()
	viper.BindPFlags(parserCmd.Flags())
}

func loop(reader io.Reader, client MQTT.Client) {
	channel := make(chan data.SensorData)
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			line := strings.SplitN(scanner.Text(), ":", 2)
			if nil == sensors.Sensors[line[0]] {
				jww.DEBUG.Println(scanner.Text())
			} else {
				d := sensors.Sensors[line[0]].Parse(line[0], line[1])
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
			jww.ERROR.Println("Failed to send message.", token.Error())
		}
		jww.DEBUG.Printf("Publishing %s -> %s\n", topic, buf.Bytes())
	}
}

func serialLoop(client MQTT.Client) {
	m := &serial.Mode{
		BaudRate: viper.GetInt("baud"),
	}
	s, err := serial.Open(viper.GetString("port"), m)
	if err != nil {
		jww.FATAL.Println(err)
		panic(err)
	}
	defer s.Close()

	s.SetDTR(true)
	s.SetRTS(true)
	time.Sleep(1 * time.Second)
	s.SetDTR(false)
	s.SetRTS(false)
	loop(s, client)
}

func parser(cmd *cobra.Command, args []string) {
	if verbose {
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	clientid := fmt.Sprintf("gowx-parser-%s-%d", hostname, os.Getpid())
	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("broker")).SetClientID(clientid).SetCleanSession(true)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		jww.FATAL.Println(token.Error())
		panic(token.Error())
	}
	defer client.Disconnect(0)

	fi, err := os.Stat(viper.GetString("port"))
	if err != nil {
		jww.FATAL.Println(err)
		panic(err)
	}
	if fi.Mode()&os.ModeType != 0 {
		serialLoop(client)
	} else {
		file, err := os.Open(viper.GetString("port"))
		if err != nil {
			jww.FATAL.Println(err)
			panic(err)
		}
		defer file.Close()
		loop(file, client)
	}
}
