// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

// manCmd represents the man command
var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate man pages",
	Long:  `Generates a set of man pages for gowx`,
	Run: func(cmd *cobra.Command, args []string) {
		header := &doc.GenManHeader{
			Title:   "GOWX",
			Section: "3",
		}
		doc.GenManTree(RootCmd, header, viper.GetString("output"))
	},
}

func init() {
	docCmd.AddCommand(manCmd)
}
