package parser

func ExampleParseTHGR122NX() {
	var o Oregon
	o.Parse("OS3", "1D20485C480882835")
	o.Parse("OS3", "1D2016B1091073A14")
	// Output:
	// Sensor ID 1D20 Channel 4 Rolling Code 85 Flags C
	// Temperature -8.4 Humidity 28
	// Sensor ID 1D20 Channel 1 Rolling Code 6B Flags 1
	// Temperature 19 Humidity 37
}
