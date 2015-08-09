package main

import (
	"github.com/geoffholden/gowx/gowx"
	"reflect"
	"testing"
)

func TestParseTHGR122NX(t *testing.T) {
	var o Oregon
	var c gowx.Config
	res := o.Parse("OS3", "1D20485C480882835", &c)
	if res.ID != "OS3:1D20" {
		t.Error("Error parsing ID")
	}
	if res.Data["Temperature"] != -8.4 {
		t.Error("Error parsing temperature")
	}
	if res.Data["Humidity"] != 28 {
		t.Error("Error parsing humidity")
	}

	res = o.Parse("OS3", "1D2016B1091073A14", &c)
	if res.Data["Temperature"] != 19 {
		t.Error("Error parsing temperature")
	}
	if res.Data["Humidity"] != 37 {
		t.Error("Error parsing humidity")
	}
}

func TestTruncated(t *testing.T) {
	var o Oregon
	var c gowx.Config

	var empty gowx.SensorData

	res := o.Parse("OS3", "1D20485C48088283", &c)
	if !reflect.DeepEqual(res, empty) {
		t.Error("SensorResult should be empty")
	}

	res = o.Parse("OS3", "1D20485C480", &c)
	if !reflect.DeepEqual(res, empty) {
		t.Error("SensorResult should be empty")
	}
}

func TestBadChecksum(t *testing.T) {
	var o Oregon
	var c gowx.Config
	var empty gowx.SensorData

	res := o.Parse("OS3", "1D20485C480882845", &c)
	if !reflect.DeepEqual(res, empty) {
		t.Error("SensorResult should be empty")
	}
}
