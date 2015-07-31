package main

import (
	"testing"
)

func TestParseTHGR122NX(t *testing.T) {
	var o Oregon
	res := o.Parse("OS3", "1D20485C480882835")
	if res.ID != "1D20" {
		t.Error("Error parsing ID")
	}
	if res.Data["Temperature"] != -8.4 {
		t.Error("Error parsing temperature")
	}
	if res.Data["Humidity"] != 28 {
		t.Error("Error parsing humidity")
	}

	res = o.Parse("OS3", "1D2016B1091073A14")
	if res.Data["Temperature"] != 19 {
		t.Error("Error parsing temperature")
	}
	if res.Data["Humidity"] != 37 {
		t.Error("Error parsing humidity")
	}
}
