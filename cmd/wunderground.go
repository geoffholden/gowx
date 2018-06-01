// Copyright © 2018 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/geoffholden/gowx/data"
	"github.com/geoffholden/gowx/units"
)

// wundergroundCmd represents the weather undergound command
var wundergroundCmd = &cobra.Command{
	Use:   "wu",
	Short: "Push updates to Weather Underground",
	Long:  `Pushes weather updates to Weather Undergound PWS`,
	Run:   wunderground,
}

func wundergroundInit() {
}

func init() {
	RootCmd.AddCommand(wundergroundCmd)
	wundergroundInit()
	viper.BindPFlags(wundergroundCmd.Flags())
}

func wunderground(cmd *cobra.Command, args []string) {
	if verbose {
		jww.SetStdoutThreshold(jww.LevelTrace)
	}

	dataChannel := make(chan aggdata)

	db, err := data.OpenDatabase()
	if err != nil {
		jww.FATAL.Println(err)
		panic(err)
	}

	topic := "/gowx/sample/aggregated"
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	clientid := fmt.Sprintf("gowx-wunderground-%s-%d", hostname, os.Getpid())
	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("broker")).SetClientID(clientid).SetCleanSession(true)

	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(topic, 0, func(client MQTT.Client, msg MQTT.Message) {
			r := bytes.NewReader(msg.Payload())
			decoder := json.NewDecoder(r)
			var data aggdata
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

	timer := time.NewTimer(5 * time.Second)
	timer.Stop()

	params := make(map[string]string)
	for {
		select {
		case <-timer.C:
			// timer expired
			if len(params) > 0 {
				sendData(params)
				// Don't wipe the old data.
				// params = make(map[string]string)
			}
		case d := <-dataChannel:
			// incoming data
			wuAddData(d, &params, db)
			timer.Stop()
			timer.Reset(5 * time.Second)
		case <-time.After(30 * time.Minute):
			jww.ERROR.Println("No data in 30 minutes, reconnecting")
			connect(client)
		}
	}
}

func sendData(params map[string]string) {
	v := url.Values{}
	for key, value := range params {
		v.Add(key, value)
	}
	v.Add("dateutc", "now")
	v.Add("action", "updateraw")
	v.Add("ID", viper.GetString("wu_id"))
	v.Add("PASSWORD", viper.GetString("wu_key"))
	res, err := http.Get("https://weatherstation.wunderground.com/weatherstation/updateweatherstation.php?" + v.Encode())
	if err != nil {
		jww.ERROR.Println(err)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		jww.ERROR.Println(err)
		return
	}
	resString := fmt.Sprintf("%s", body)
	if resString != "success\n" {
		jww.ERROR.Println(resString)
	}
}

func wuAddData(d aggdata, params *map[string]string, db *data.Database) {
	config := viper.Sub("wunderground")
	rxp := regexp.MustCompile(`\[([^]]*)\]`)

	for elem, _ := range viper.GetStringMap("wunderground") {
		query := config.GetStringMapString(elem)

		key := rxp.ReplaceAllString(d.Key.Key, "") // drop any modifiers

		if x, ok := query["type"]; ok {
			if key != x {
				continue
			}
		}

		if x, ok := query["id"]; ok {
			if d.Key.ID != x {
				continue
			}
		}

		if x, ok := query["channel"]; ok {
			if strconv.FormatInt(int64(d.Key.Channel), 10) != x {
				continue
			}
		}
		if x, ok := query["serial"]; ok {
			if d.Key.Serial != x {
				continue
			}
		}

		// We have a match!
		col := rxp.FindStringSubmatch(d.Key.Key)
		value := d.Avg
		if len(col) > 1 {
			switch col[1] {
			case "min":
				value = d.Min
			case "max":
				value = d.Max
			}
		}

		// Do any unit conversion here, also special case for rain totals
		switch elem {
		case "windspeedmph", "windgustmph", "windspdmph_avg2m", "windgustmph_10m":
			u := units.NewSpeedMetersPerSecond(value)
			value = u.MilesPerHour()
		case "dewptf", "tempf", "soiltempf":
			u := units.NewTemperatureCelsius(value)
			value = u.Fahrenheit()
		case "rainin":
			t := time.Now().UTC().Unix() - 3600
			old, err := db.QueryFirst(t, d.Key.Key, d.Key.ID, d.Key.Channel)
			if err != nil {
				jww.ERROR.Println(err)
				continue
			}
			value = value - old
			u := units.NewDistanceMillimeters(value)
			value = u.Inches()
		case "dailyrainin":
			t := bod(time.Now()).UTC().Unix()
			old, err := db.QueryFirst(t, d.Key.Key, d.Key.ID, d.Key.Channel)
			if err != nil {
				jww.ERROR.Println(err)
				continue
			}
			value = value - old
			u := units.NewDistanceMillimeters(value)
			value = u.Inches()
		case "baromin":
			u := units.NewPressureHectopascal(value)
			value = u.InchMercury()
		case "visibility":
			u := units.NewDistanceMeters(value)
			value = u.NauticalMiles()
		}

		(*params)[elem] = strconv.FormatFloat(value, 'f', -1, 64)
	}
}

func bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

/*
winddir - [0-360 instantaneous wind direction]
windspeedmph - [mph instantaneous wind speed]
windgustmph - [mph current wind gust, using software specific time period]
windgustdir - [0-360 using software specific time period]
windspdmph_avg2m  - [mph 2 minute average wind speed mph]
winddir_avg2m - [0-360 2 minute average wind direction]
windgustmph_10m - [mph past 10 minutes wind gust mph ]
windgustdir_10m - [0-360 past 10 minutes wind gust direction]
humidity - [% outdoor humidity 0-100%]
dewptf- [F outdoor dewpoint F]
tempf - [F outdoor temperature]
rainin - [rain inches over the past hour)] -- the accumulated rainfall in the past 60 min
dailyrainin - [rain inches so far today in local time]
baromin - [barometric pressure inches]
weather - [text] -- metar style (+RA)
clouds - [text] -- SKC, FEW, SCT, BKN, OVC
soiltempf - [F soil temperature]
soilmoisture - [%]
leafwetness  - [%]
solarradiation - [W/m^2]
UV - [index]
visibility - [nm visibility]
*/
