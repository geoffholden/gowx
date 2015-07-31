package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"regexp"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/geoffholden/gowx/gowx"
	_ "github.com/mattn/go-sqlite3"
)

type mapKey struct {
	ID      string
	Channel int
	Serial  string
	Key     string
}

type Aggregator struct {
	config gowx.Config
	data   map[mapKey][]float64
	db     *sql.DB
}

func main() {
	flag.Parse()
	var aggregator Aggregator
	aggregator.config = gowx.NewConfig().LoadFile()

	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}
	aggregator.db = db
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS samples (
		timestamp   integer,
		id          text,
		channel     integer,
		serial      string,
		key         string,
		min         real,
		max         real,
		avg         real
	)`)
	if err != nil {
		panic(err)
	}

	dataChannel := make(chan gowx.SensorData)

	topic := "/gowx/sample"
	opts := MQTT.NewClientOptions().AddBroker(aggregator.config.Global.MQTTServer).SetClientID("aggregator").SetCleanSession(true)

	opts.OnConnect = func(c *MQTT.Client) {
		if token := c.Subscribe(topic, 0, func(client *MQTT.Client, msg MQTT.Message) {
			r := bytes.NewReader(msg.Payload())
			decoder := json.NewDecoder(r)
			var data gowx.SensorData
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

	ticker := time.NewTicker(time.Duration(aggregator.config.Aggregator.AverageInterval) * time.Second)

	aggregator.data = make(map[mapKey][]float64)
	for {
		select {
		case <-ticker.C:
			aggregator.sumData()
		case d := <-dataChannel:
			aggregator.addData(d)
		}
	}
}

func (a *Aggregator) addData(data gowx.SensorData) {
	fmt.Printf("Adding data:\n")
	for k, v := range data.Data {
		fmt.Printf("\t\t%s -> %f\n", k, v)
		key := mapKey{
			ID:      data.ID,
			Channel: data.Channel,
			Serial:  data.Serial,
			Key:     k,
		}
		a.data[key] = append(a.data[key], v)
	}
}

func (a *Aggregator) sumData() {
	direction := regexp.MustCompile(`Dir$`)

	stmt := `INSERT INTO samples (
		timestamp,
		id,
		channel,
		serial,
		key,
		min, max, avg
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	timestamp := time.Now().UTC().Unix()

	for key, slice := range a.data {
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
		_, err := a.db.Exec(stmt, timestamp, key.ID, key.Channel, key.Serial, key.Key, min, max, avg)
		if err != nil {
			fmt.Errorf("%s\n", err.Error())
		}
	}
	a.data = make(map[mapKey][]float64)
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
