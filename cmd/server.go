// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/geoffholden/gowx/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"web", "serve", "webserver"},
	Short:   "Web Server",
	Long:    `Launches the web server.`,
	Run:     server,
}

func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.Flags().String("webroot", "web", "Root directory for the web server.")
	serverCmd.Flags().String("address", ":0", "Address and port to listen on.")
	viper.BindPFlags(serverCmd.Flags())
}

func server(cmd *cobra.Command, args []string) {
	var currentData struct {
		Temperature float64
		Humidity    float64
		Pressure    float64
		Wind        float64
		WindDir     float64
		RainRate    float64
	}

	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("broker")).SetClientID("web").SetCleanSession(true)
	opts.OnConnect = func(c *MQTT.Client) {
		if token := c.Subscribe("/gowx/sample", 0, func(client *MQTT.Client, msg MQTT.Message) {
			r := bytes.NewReader(msg.Payload())
			decoder := json.NewDecoder(r)
			var data data.SensorData
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
	db, err := data.OpenDatabase()
	if err != nil {
		panic(err)
	}

	http.Handle("/", http.FileServer(http.Dir(viper.GetString("webroot"))))

	http.HandleFunc("/data.json", func(w http.ResponseWriter, r *http.Request) {
		dataHandler(w, r, db)
	})

	http.HandleFunc("/change.json", func(w http.ResponseWriter, r *http.Request) {
		changeHandler(w, r, db)
	})

	http.HandleFunc("/wind.json", func(w http.ResponseWriter, r *http.Request) {
		windHandler(w, r, db)
	})

	http.HandleFunc("/currentdata.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(currentData)
	})

	listener, err := net.Listen("tcp", viper.GetString("address"))
	if err != nil {
		panic(err)
	}
	addr := listener.Addr()
	fmt.Println("Listening on", addr.String())

	http.Serve(listener, nil)
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

func dataHandler(w http.ResponseWriter, r *http.Request, db *data.Database) {
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
			var rows <-chan data.Row
			if interval > 1 {
				rows = db.QueryRowsInterval(t, key, id, interval)
			} else {
				rows = db.QueryRows(t, key, id)
			}

			col := rxp.FindStringSubmatch(datatype)

			for row := range rows {
				_, off := time.Unix(row.Timestamp, 0).Zone()
				t := (time.Unix(row.Timestamp, 0).Unix() + int64(off)) * 1000
				sub := make([]interface{}, 2)
				sub[0] = t
				if len(col) > 1 {
					switch col[1] {
					case "min":
						sub[1] = row.Min
					case "max":
						sub[1] = row.Max
					case "avg":
						sub[1] = row.Avg
					default:
						sub[1] = row.Avg
					}
				} else {
					sub[1] = row.Avg
				}
				result.Data[index] = append(result.Data[index], sub)

				sub = make([]interface{}, 3)
				sub[0] = t
				sub[1] = row.Min
				sub[2] = row.Max
				result.Errorbars[index] = append(result.Errorbars[index], sub)
			}
			index++
		}
	}
	json.NewEncoder(w).Encode(result)
}

func windHandler(w http.ResponseWriter, r *http.Request, db *data.Database) {
	t, _ := computeTime(r.FormValue("time"))

	var result struct {
		Wind []float64
		Gust []float64
	}

	result.Wind = make([]float64, 32)
	result.Gust = make([]float64, 32)

	for row := range db.QueryWind(t) {
		result.Wind[int(row.Dir)] = row.Avg
		result.Gust[int(row.Dir)] = row.Gust
	}
	json.NewEncoder(w).Encode(result)
}

func changeHandler(w http.ResponseWriter, r *http.Request, db *data.Database) {
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
			old, err := db.QueryFirst(t, datatype, id, channel)
			if err != nil {
			}

			now, err := db.QueryLast(t, datatype, id, channel)
			if err != nil {
			}

			result.Change = append(result.Change, now-old)

			index++
		}
	}
	json.NewEncoder(w).Encode(result)
}
