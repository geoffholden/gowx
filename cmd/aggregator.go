// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/geoffholden/gowx/data"
)

type aggdata struct {
	Timestamp int64
	Key       mapKey
	Min       float64
	Max       float64
	Avg       float64
}

// aggregatorCmd represents the aggregator command
var aggregatorCmd = &cobra.Command{
	Use:   "aggregator",
	Short: "Aggregates individual samples",
	Long:  `Aggregates individual samples and stores the result into the database.`,
	Run:   aggregator,
}

func aggregatorInit() {
	if !aggregatorCmd.Flags().HasFlags() {
		aggregatorCmd.Flags().Int("interval", 300, "Interval (in seconds) to aggregate data.")
	}
}

func init() {
	RootCmd.AddCommand(aggregatorCmd)
	aggregatorInit()
	viper.BindPFlags(aggregatorCmd.Flags())
}

type mapKey struct {
	ID      string
	Channel int
	Serial  string
	Key     string
}

func aggregator(cmd *cobra.Command, args []string) {
	if verbose {
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
	db, err := data.OpenDatabase()
	if err != nil {
		jww.FATAL.Println(err)
		panic(err)
	}

	dataChannel := make(chan data.SensorData)

	topic := "/gowx/sample"
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	clientid := fmt.Sprintf("gowx-aggregator-%s-%d", hostname, os.Getpid())
	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("broker")).SetClientID(clientid).SetCleanSession(true)

	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(topic, 0, func(client MQTT.Client, msg MQTT.Message) {
			r := bytes.NewReader(msg.Payload())
			decoder := json.NewDecoder(r)
			var data data.SensorData
			err := decoder.Decode(&data)
			if err != nil {
				jww.ERROR.Println(err)
				return
			}
			dataChannel <- data
		}); token.Wait() && token.Error() != nil {
			jww.FATAL.Println(token.Error())
			panic(token.Error())
		}
	}

	opts.OnConnectionLost = func(c MQTT.Client, e error) {
		jww.ERROR.Println("MQTT Connection Lost", e)
		if token := c.Connect(); token.Wait() && token.Error() != nil {
			jww.FATAL.Println(token.Error())
			panic(token.Error())
		}
	}

	opts.AutoReconnect = false

	client := MQTT.NewClient(opts)
	connect(client)
	defer client.Disconnect(0)

	ticker := time.NewTicker(time.Duration(viper.GetInt("interval")) * time.Second)

	thedata := make(map[mapKey][]float64)
	for {
		select {
		case <-ticker.C:
			res := sumData(&thedata, db)
			publishData(res, db, client)
		case d := <-dataChannel:
			addData(&thedata, d)
		case <-time.After(5 * time.Minute):
			jww.ERROR.Println("No data in 5 minutes, reconnecting")
			connect(client)
		}
	}
}

func addData(thedata *map[mapKey][]float64, d data.SensorData) {
	if len(d.Data) > 0 {
		jww.DEBUG.Printf("Adding data:\n")
	}
	for k, v := range d.Data {
		jww.DEBUG.Printf("\t\t%s -> %f\n", k, v)
		key := mapKey{
			ID:      d.ID,
			Channel: d.Channel,
			Serial:  d.Serial,
			Key:     k,
		}
		(*thedata)[key] = append((*thedata)[key], v)
	}
}

func sumData(thedata *map[mapKey][]float64, db *data.Database) []aggdata {
	direction := regexp.MustCompile(`Dir$`)

	timestamp := time.Now().UTC().Unix()

	result := make([]aggdata, len(*thedata))
	index := 0

	for key, slice := range *thedata {
		var min, max, avg float64
		switch {
		case direction.MatchString(key.Key):
			min = circularmean(slice)
			max = min
			avg = min
		default:
			min = minimum(slice)
			max = maximum(slice)
			avg = mean(slice)
		}

		result[index].Timestamp = timestamp
		result[index].Key = key
		result[index].Min = min
		result[index].Max = max
		result[index].Avg = avg
		index++
	}
	*thedata = make(map[mapKey][]float64)
	return result
}

func publishData(data []aggdata, db *data.Database, client MQTT.Client) {
	for _, d := range data {
		// publish the data to the database
		err := db.InsertRow(d.Timestamp, d.Key.ID, d.Key.Channel, d.Key.Serial, d.Key.Key, d.Min, d.Max, d.Avg)
		if err != nil {
			jww.ERROR.Printf("%s\n", err.Error())
		}

		// publish the data to the broker
		topic := "/gowx/sample/aggregated"
		buf := new(bytes.Buffer)
		encoder := json.NewEncoder(buf)
		encoder.Encode(d)
		payload := buf.Bytes()
		if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			jww.ERROR.Println("Failed to send message.", token.Error())
		}
	}
}

func minimum(d []float64) float64 {
	result := d[0]
	for _, x := range d {
		if x < result {
			result = x
		}
	}
	return result
}

func maximum(d []float64) float64 {
	result := d[0]
	for _, x := range d {
		if x > result {
			result = x
		}
	}
	return result
}

func mean(d []float64) float64 {
	sum := 0.0
	for _, x := range d {
		sum += x
	}
	return sum / float64(len(d))
}

func circularmean(d []float64) float64 {
	var sumsin, sumcos float64

	for _, x := range d {
		rad := x * math.Pi / 180.0
		sumsin += math.Sin(rad)
		sumcos += math.Cos(rad)
	}
	return math.Atan2(sumsin/float64(len(d)), sumcos/float64(len(d))) * 180.0 / math.Pi
}
