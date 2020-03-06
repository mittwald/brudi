package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	author = "Dennis Hermsmeier <Mittwald CM Service>"
	url    = "https://github.com/mittwald/brudi"
	commit = ""
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of brudi",
	Long:  `All software has versions. This is brudi's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(
			""+
				"\n"+
				"Commit: %s\n"+
				"Author: %s\n"+
				"Visit: %s\n"+
				"",
			commit, author, url)
	},
}
