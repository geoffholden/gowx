// Copyright © 2017 Geoff Holden <geoff@geoffholden.com>

package sensors

import (
	"github.com/geoffholden/gowx/data"
	"reflect"
	"testing"
)

func TestParseVN1TX(t *testing.T) {
	var v VN1TX
	res := v.Parse("VN1", "E6D271006F00AC446")
	if res.ID != "VN1:6D27" {
		t.Error("Error parsing ID")
	}
	if res.Data["CurrentWind"] < 1.65 || res.Data["CurrentWind"] > 1.66 {
		t.Error("Error parsing wind speed", res.Data["CurrentWind"])
	}
	if res.Data["WindDir"] != 180 {
		t.Error("Error parsing wind direction")
	}
	if res.Data["RainTotal"] != 11.176 {
		t.Error("Error parsing rain total")
	}

	res = v.Parse("VN1", "E6D27800C665DB366")
	if res.Data["Temperature"] < 8.277 || res.Data["Temperature"] > 8.278 {
		t.Error("Error parsing temperature")
	}
	if res.Data["Humidity"] != 91 {
		t.Error("Error parsing humidity")
	}
}

func TestTruncated(t *testing.T) {
	var v VN1TX

	var empty data.SensorData

	res := v.Parse("VN1", "E6D27800C665DB36")
	if !reflect.DeepEqual(res, empty) {
		t.Error("SensorResult should be empty")
	}

	res = v.Parse("VN1", "E6D27800")
	if !reflect.DeepEqual(res, empty) {
		t.Error("SensorResult should be empty")
	}
}

func TestBadChecksum(t *testing.T) {
	var v VN1TX
	var empty data.SensorData

	res := v.Parse("VN1", "E6D271006F00AC456")
	if !reflect.DeepEqual(res, empty) {
		t.Error("SensorResult should be empty")
	}
}
