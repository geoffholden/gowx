// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package data

import "time"

type SensorData struct {
	TimeStamp time.Time
	ID        string
	Channel   int
	Serial    string
	Data      map[string]float64
}

type SensorParser interface {
	Parse(key string, data string) SensorData
}
