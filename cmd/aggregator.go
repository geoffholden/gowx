// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/geoffholden/gowx/data"
)

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
	fmt.Println("aggregator called")

	db, err := data.OpenDatabase()
	if err != nil {
		panic(err)
	}

	dataChannel := make(chan data.SensorData)

	topic := "/gowx/sample"
	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("broker")).SetClientID("aggregator").SetCleanSession(true)

	opts.OnConnect = func(c *MQTT.Client) {
		if token := c.Subscribe(topic, 0, func(client *MQTT.Client, msg MQTT.Message) {
			r := bytes.NewReader(msg.Payload())
			decoder := json.NewDecoder(r)
			var data data.SensorData
			err := decoder.Decode(&data)
			if err != nil {
				panic(err)
			}
			dataChannel <- data
		}); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(0)

	ticker := time.NewTicker(time.Duration(viper.GetInt("interval")) * time.Second)

	thedata := make(map[mapKey][]float64)
	for {
		select {
		case <-ticker.C:
			sumData(&thedata, db)
		case d := <-dataChannel:
			addData(&thedata, d)
		}
	}
}

func addData(thedata *map[mapKey][]float64, d data.SensorData) {
	fmt.Printf("Adding data:\n")
	for k, v := range d.Data {
		fmt.Printf("\t\t%s -> %f\n", k, v)
		key := mapKey{
			ID:      d.ID,
			Channel: d.Channel,
			Serial:  d.Serial,
			Key:     k,
		}
		(*thedata)[key] = append((*thedata)[key], v)
	}
}

func sumData(thedata *map[mapKey][]float64, db *data.Database) {
	direction := regexp.MustCompile(`Dir$`)

	timestamp := time.Now().UTC().Unix()

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
		err := db.InsertRow(timestamp, key.ID, key.Channel, key.Serial, key.Key, min, max, avg)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
	}
	*thedata = make(map[mapKey][]float64)
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
