package gowx

import (
	"code.google.com/p/gcfg"
	"flag"
)

var configFile string

type Config struct {
	Global struct {
		MQTTServer string
	}

	Parser struct {
		SerialPort string
		SerialBaud int
		Elevation  int
	}

	Aggregator struct {
		AverageInterval int
	}
}

func NewConfig() *Config {
	var c Config
	c.Global.MQTTServer = "tcp://localhost:1883"

	c.Aggregator.AverageInterval = 300

	return &c
}

func init() {
	flag.StringVar(&configFile, "config", "", "Config file to load")
}

func (c *Config) LoadFile() Config {
	if configFile != "" {
		gcfg.ReadFileInto(c, configFile)
	}
	return *c
}
