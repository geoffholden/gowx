// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

// markdownCmd represents the markdown command
var markdownCmd = &cobra.Command{
	Use:   "markdown",
	Short: "Generate Markdown documentation",
	Long:  `Generates documentation for gowx in Markdown format.`,
	Run: func(cmd *cobra.Command, args []string) {
		doc.GenMarkdownTree(RootCmd, viper.GetString("output"))
	},
}

func init() {
	docCmd.AddCommand(markdownCmd)
}
