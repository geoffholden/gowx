// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// bashCmd represents the bash command
var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate Bash autocompletion file",
	Long:  `Generates an autocompletion file for Bash`,
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletionFile(viper.GetString("output") + "gowx_completions.sh")
	},
}

func init() {
	docCmd.AddCommand(bashCmd)
}
