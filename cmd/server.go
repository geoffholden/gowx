// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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

func serverInit() {
	if !serverCmd.Flags().HasFlags() {
		serverCmd.Flags().String("webroot", "web", "Root directory for the web server.")
		serverCmd.Flags().String("address", ":0", "Address and port to listen on.")
	}
}

func init() {
	RootCmd.AddCommand(serverCmd)
	serverInit()
	viper.BindPFlags(serverCmd.Flags())

	viper.SetDefault("units", map[string]string{
		"Temperature":  "C",
		"Pressure":     "hPa",
		"RainfallRate": "mm/h",
		"RainTotal":    "mm",
		"WindSpeed":    "m/s",
	})
	viper.SetDefault("temperature", []map[string]string{{"type": "Temperature", "label": "Temperature"}})
	viper.SetDefault("pressure", []map[string]string{{"type": "Pressure", "label": "Pressure"}})
	viper.SetDefault("humidity", []map[string]string{{"type": "Humidity", "label": "Humidity"}})
	viper.SetDefault("wind", []map[string]string{{"type": "AverageWind[avg]", "label": "Average Wind"}, {"type": "CurrentWind[max]", "label": "Gusts"}})
	viper.SetDefault("rain", []map[string]string{{"type": "RainRate", "label": "Rainfall Rate"}, {"type": "RainTotal", "label": "Total Rain"}})
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

	var d templateData

	d.Units = viper.GetStringMapString("units")
	bytes, err := json.Marshal(convertToStringMap("temperature"))
	if err == nil {
		d.Temperature = string(bytes)
	}
	bytes, err = json.Marshal(convertToStringMap("pressure"))
	if err == nil {
		d.Pressure = string(bytes)
	}
	bytes, err = json.Marshal(convertToStringMap("humidity"))
	if err == nil {
		d.Humidity = string(bytes)
	}
	bytes, err = json.Marshal(convertToStringMap("wind"))
	if err == nil {
		d.Wind = string(bytes)
	}
	bytes, err = json.Marshal(convertToStringMap("rain"))
	if err == nil {
		d.Rain = string(bytes)
	}

	staticServer := http.FileServer(http.Dir(viper.GetString("webroot")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, r, staticServer, d)
	})

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

func convertToStringMap(name string) []map[string]string {
	if result, ok := viper.Get(name).([]map[string]string); ok {
		return result
	} else if array, ok := viper.Get(name).([]interface{}); ok {
		result := make([]map[string]string, len(array))
		for index, query := range array {
			if interfaceMap, ok := query.(map[interface{}]interface{}); ok {
				result[index] = make(map[string]string)
				for k, v := range interfaceMap {
					if kstr, ok := k.(string); ok {
						if vstr, ok := v.(string); ok {
							result[index][kstr] = vstr
						}
					}
				}
			}
		}
		return result
	}
	return nil
}

type templateData struct {
	Units        map[string]string
	Temperature  string
	Pressure     string
	Humidity     string
	Wind         string
	Rain         string
	TemperatureQ string
}

func serveTemplate(w http.ResponseWriter, r *http.Request, static http.Handler, thedata templateData) {
	temp, err := template.ParseGlob(viper.GetString("webroot") + "/*.html")
	if err != nil {
		panic(err)
	}
	name := r.URL.Path
	if name == "/" {
		name = "index.html"
	}
	if temp.Lookup(name) != nil {
		err = temp.ExecuteTemplate(w, name, thedata)
	} else {
		static.ServeHTTP(w, r)
	}
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
	var queries []map[string]string
	err := json.Unmarshal([]byte(r.FormValue("query")), &queries)
	if err != nil {
		fmt.Println(err)
	}
	//channel := r.FormValue("channel")

	t, interval := computeTime(r.FormValue("time"))

	var result struct {
		Data      [][]interface{}
		Errorbars [][]interface{}
		Label     []string
	}
	result.Data = make([][]interface{}, len(queries))
	result.Errorbars = make([][]interface{}, len(queries))
	result.Label = make([]string, len(queries))

	rxp := regexp.MustCompile(`\[([^]]*)\]`)
	for index, querymap := range queries {
		id := "%"
		if _, ok := querymap["id"]; ok {
			id = querymap["id"]
		}
		datatype := "%"
		if _, ok := querymap["type"]; ok {
			datatype = querymap["type"]
		}
		if _, ok := querymap["label"]; ok {
			result.Label[index] = querymap["label"]
		} else {
			result.Label[index] = "Unknown"
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
	json.NewEncoder(w).Encode(result)
}

func windHandler(w http.ResponseWriter, r *http.Request, db *data.Database) {
	t, _ := computeTime(r.FormValue("time"))

	var queries []map[string]string
	err := json.Unmarshal([]byte(r.FormValue("query")), &queries)
	if err != nil {
		fmt.Println(err)
	}

	var result struct {
		Data  [][]float64
		Label []string
	}

	result.Data = make([][]float64, len(queries))
	result.Label = make([]string, len(queries))

	rxp := regexp.MustCompile(`\[([^]]*)\]`)
	for index, querymap := range queries {
		result.Label[index] = querymap["label"]
		result.Data[index] = make([]float64, 32)

		id := "%"
		if _, ok := querymap["id"]; ok {
			id = querymap["id"]
		}
		datatype := "%"
		if _, ok := querymap["type"]; ok {
			datatype = querymap["type"]
		}
		if _, ok := querymap["label"]; ok {
			result.Label[index] = querymap["label"]
		} else {
			result.Label[index] = "Unknown"
		}
		cols := rxp.FindStringSubmatch(datatype)
		var col string
		if len(cols) > 1 {
			col = cols[1]
		} else {
			col = "avg"
		}

		key := rxp.ReplaceAllString(datatype, "")

		for row := range db.QueryWind(t, key, col, id, 0) {
			result.Data[index][int(row.Dir)] = row.Value
		}
	}

	// result.Wind = make([]float64, 32)
	// result.Gust = make([]float64, 32)

	// for row := range db.QueryWind(t) {
	// 	result.Wind[int(row.Dir)] = row.Avg
	// 	result.Gust[int(row.Dir)] = row.Gust
	// }
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
