// Copyright © 2016 Geoff Holden (geoff@geoffholden.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/geoffholden/gowx/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tarm/serial"

	"github.com/geoffholden/gowx/sensors"
)

// parserCmd represents the parser command
var parserCmd = &cobra.Command{
	Use:   "parser",
	Short: "Parse serial data",
	Long:  `Parses the serial data coming from the WSDL sheild and sends it to the MQTT broker.`,
	Run:   parser,
}

func init() {
	RootCmd.AddCommand(parserCmd)

	parserCmd.Flags().String("port", "", "Serial port to connect to")
	parserCmd.Flags().Int("baud", 9600, "Serial baud rate")
	parserCmd.Flags().Int("elevation", 0, "Elevation about sea level")

	viper.BindPFlags(parserCmd.Flags())
}

func loop(reader io.Reader, client *MQTT.Client) {
	channel := make(chan data.SensorData)
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			line := strings.SplitN(scanner.Text(), ":", 2)
			if nil == sensors.Sensors[line[0]] {
				fmt.Println(scanner.Text())
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
			fmt.Println("Failed to send message.")
			panic(token.Error())
		}
		fmt.Printf("Publishing %s -> %s\n", topic, buf.Bytes())
	}
}

func serialLoop(client *MQTT.Client) {
	c := &serial.Config{
		Name: viper.GetString("port"),
		Baud: viper.GetInt("baud"),
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
	defer s.Close()
	s.Flush()
	loop(s, client)
}

func parser(cmd *cobra.Command, args []string) {
	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("broker")).SetClientID("parser").SetCleanSession(true)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(0)

	fi, err := os.Stat(viper.GetString("port"))
	if err != nil {
		panic(err)
	}
	if fi.Mode()&os.ModeType != 0 {
		serialLoop(client)
	} else {
		file, err := os.Open(viper.GetString("port"))
		if err != nil {
			panic(err)
		}
		defer file.Close()
		loop(file, client)
	}
}
