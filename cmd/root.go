// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"os"
	"strings"

	"github.com/geoffholden/gowx/data"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool

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
		jww.ERROR.Println(err)
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
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	dbdrivers := data.DBDrivers()
	if len(dbdrivers) > 1 {
		RootCmd.PersistentFlags().String("dbDriver", "sqlite3", "Database Driver, one of ["+strings.Join(dbdrivers, ", ")+"]")
	} else {
		viper.SetDefault("dbDriver", "sqlite3")
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
		jww.DEBUG.Println("Using config file:", viper.ConfigFileUsed())
	}
}
