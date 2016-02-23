// Copyright © 2016 Geoff Holden (geoff@geoffholden.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/geoffholden/gowx/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gowx",
	Short: "Go Weather Station Interface",
	Long: `Go Weather Station Interface is a data logger and web interface for
weather station data.

It currently parses the serial data from the WSDL WxShield for Arduino,
store the parsed data in a database, and has a web display interface.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is gowx.yaml)")
	RootCmd.PersistentFlags().String("broker", "tcp://localhost:1883", "MQTT Server")
	RootCmd.PersistentFlags().String("database", "gowx.db", "Database")

	dbdrivers := data.DBDrivers()
	if len(dbdrivers) > 1 {
		RootCmd.PersistentFlags().String("dbDriver", "sqlite3", "Database Driver, one of ["+strings.Join(dbdrivers, ", ")+"]")
	}
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	viper.BindPFlags(RootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("gowx") // name of config file (without extension)
	viper.AddConfigPath("/etc/gowx/")
	viper.AddConfigPath("$HOME/.gowx/")
	viper.AddConfigPath(".")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
