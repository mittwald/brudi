package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	author = "Mittwald CM Service <https://github.com/mittwald/brudi>"
	tag    = "dev"
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
				"Tag: %s\n"+
				"Author: %s\n"+
				"",
			tag, author)
	},
}
