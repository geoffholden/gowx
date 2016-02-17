package gowx

import (
	"flag"
	"gopkg.in/gcfg.v1"
)

var configFile string

type Config struct {
	Global struct {
		MQTTServer     string
		DatabaseDriver string
		Database       string
	}

	Parser struct {
		SerialPort string
		SerialBaud int
		Elevation  int
	}

	Aggregator struct {
		AverageInterval int
	}

	Web struct {
		Root string
	}
}

func NewConfig() *Config {
	var c Config
	c.Global.MQTTServer = "tcp://localhost:1883"

	c.Global.DatabaseDriver = "sqlite3"
	c.Global.Database = "gowx.db"

	c.Aggregator.AverageInterval = 300

	c.Web.Root = "./webroot"

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
