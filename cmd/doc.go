// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// docCmd represents the doc command
var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Documentation generator",
	Long:  `Generators for documentation and shell completion.`,
}

func init() {
	RootCmd.AddCommand(docCmd)

	docCmd.PersistentFlags().String("output", "./", "Output directory")
	viper.BindPFlags(docCmd.PersistentFlags())
}
