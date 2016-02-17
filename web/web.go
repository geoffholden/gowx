package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/geoffholden/gowx/gowx"
	_ "github.com/mattn/go-sqlite3"
)

var currentData struct {
	Temperature float64
	Humidity    float64
	Pressure    float64
	Wind        float64
	WindDir     float64
	RainRate    float64
}

var db *sql.DB

func main() {
	flag.Parse()

	config := gowx.NewConfig().LoadFile()

	opts := MQTT.NewClientOptions().AddBroker(config.Global.MQTTServer).SetClientID("web").SetCleanSession(true)
	opts.OnConnect = func(c *MQTT.Client) {
		if token := c.Subscribe("/gowx/sample", 0, func(client *MQTT.Client, msg MQTT.Message) {
			r := bytes.NewReader(msg.Payload())
			decoder := json.NewDecoder(r)
			var data gowx.SensorData
			err := decoder.Decode(&data)
			if err != nil {
				panic(err)
			}
			switch data.ID {
			case "OS3:F824":
				currentData.Temperature = data.Data["Temperature"]
				currentData.Humidity = data.Data["Humidity"]
			case "BMP":
				currentData.Pressure = data.Data["Pressure"]
			case "OS3:1984":
				currentData.Wind = data.Data["AverageWind"]
				currentData.WindDir = data.Data["WindDir"]
			case "OS3:2914":
				currentData.RainRate = data.Data["RainRate"]
			}
		}); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(0)

	var err error
	db, err = sql.Open(config.Global.DatabaseDriver, config.Global.Database)
	if err != nil {
		panic(err)
	}

	http.Handle("/", http.FileServer(http.Dir(config.Web.Root)))

	http.HandleFunc("/data.json", data)

	http.HandleFunc("/change.json", change)

	http.HandleFunc("/wind.json", wind)

	http.HandleFunc("/currentdata.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(currentData)
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func computeTime(timestr string) (int64, int64) {
	t := time.Now().UTC().Unix()
	interval := int64(1)

	rxp := regexp.MustCompile(`^([0-9]+)([hd])$`)
	if !rxp.MatchString(timestr) {
		timestr = "24h"
	}
	match := rxp.FindStringSubmatch(timestr)

	var val, mult int64
	val, _ = strconv.ParseInt(match[1], 10, 64)
	switch match[2] {
	case "h":
		mult = 60 * 60
	case "d":
		mult = 60 * 60 * 24
	default:
		mult = 60 * 60 * 24
	}

	td := val * mult

	if td > 60*60*24*30 {
		interval = 12 * 60 * 60
	} else if td > 60*60*24*7 {
		interval = 2 * 60 * 60
	} else if td > 60*60*24 {
		interval = 30 * 60
	}

	return t - td, interval
}

func data(w http.ResponseWriter, r *http.Request) {
	datatypes := strings.Split(r.FormValue("type"), ",")
	ids := strings.Split(r.FormValue("id"), ",")
	//channel := r.FormValue("channel")

	t, interval := computeTime(r.FormValue("time"))

	var result struct {
		Data      [][]interface{}
		Errorbars [][]interface{}
	}
	result.Data = make([][]interface{}, len(datatypes)*len(ids))
	result.Errorbars = make([][]interface{}, len(datatypes)*len(ids))

	rxp := regexp.MustCompile(`\[([^]]*)\]`)
	index := 0
	for _, datatype := range datatypes {
		for _, id := range ids {
			if id == "" {
				id = "%"
			}
			key := rxp.ReplaceAllString(datatype, "")
			var rows *sql.Rows
			if interval > 1 {
				rows, _ = db.Query("select cast(timestamp/? as INTEGER) * ? as ts,min(min),max(max),avg(avg) from samples where key = ? and id like ? and timestamp > ? group by ts order by ts", interval, interval, key, id, t)
			} else {
				rows, _ = db.Query("select timestamp,min,max,avg from samples where key = ? and id like ? and timestamp > ? order by timestamp", key, id, t)
			}
			defer rows.Close()

			col := rxp.FindStringSubmatch(datatype)

			for rows.Next() {
				var timestamp int64
				var min, max, avg float64
				err := rows.Scan(&timestamp, &min, &max, &avg)
				if err != nil {
					panic(err)
				}
				_, off := time.Unix(timestamp, 0).Zone()
				t := (time.Unix(timestamp, 0).Unix() + int64(off)) * 1000
				sub := make([]interface{}, 2)
				sub[0] = t
				if len(col) > 1 {
					switch col[1] {
					case "min":
						sub[1] = min
					case "max":
						sub[1] = max
					case "avg":
						sub[1] = avg
					default:
						sub[1] = avg
					}
				} else {
					sub[1] = avg
				}
				result.Data[index] = append(result.Data[index], sub)

				sub = make([]interface{}, 3)
				sub[0] = t
				sub[1] = min
				sub[2] = max
				result.Errorbars[index] = append(result.Errorbars[index], sub)
			}
			index++
		}
	}
	json.NewEncoder(w).Encode(result)
}

func wind(w http.ResponseWriter, r *http.Request) {
	t, _ := computeTime(r.FormValue("time"))

	var result struct {
		Wind []float64
		Gust []float64
	}

	result.Wind = make([]float64, 32)
	result.Gust = make([]float64, 32)

	rows, err := db.Query("select (((dir.avg + 5.125 + 360) % 360) / 11.25) % 32 as dir,max(gust.max),avg(avg.avg) from samples dir inner join samples gust on gust.timestamp = dir.timestamp inner join samples avg on avg.timestamp = gust.timestamp where dir.key = 'WindDir' and gust.key = 'CurrentWind' and avg.key = 'AverageWind' and dir.timestamp > ? group by dir;", t)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var dir, gust, avg float64
		err := rows.Scan(&dir, &gust, &avg)
		if err != nil {
			panic(err)
		}
		result.Wind[int(dir)] = avg
		result.Gust[int(dir)] = gust
	}
	json.NewEncoder(w).Encode(result)
}

func change(w http.ResponseWriter, r *http.Request) {
	datatypes := strings.Split(r.FormValue("type"), ",")
	ids := strings.Split(r.FormValue("id"), ",")
	channel, err := strconv.Atoi(r.FormValue("channel"))
	if err != nil {
		channel = 0
	}

	t, _ := computeTime(r.FormValue("time"))

	var result struct {
		Change []float64
	}

	index := 0
	for _, datatype := range datatypes {
		for _, id := range ids {
			if id == "" {
				id = "%"
			}
			row := db.QueryRow("select avg from samples where key = ? and id like ? and channel = ? and timestamp > ? order by timestamp limit 1", datatype, id, channel, t)

			var old float64
			err := row.Scan(&old)
			if err != nil {
			}

			row = db.QueryRow("select avg from samples where key = ? and id like ? and channel = ? and timestamp > ? order by timestamp desc limit 1", datatype, id, channel, t)

			var now float64
			err = row.Scan(&now)
			if err != nil {
			}

			result.Change = append(result.Change, now-old)

			index++
		}
	}
	json.NewEncoder(w).Encode(result)
}
