package main

import (
	"github.com/spf13/cobra"

	"github.com/mittwald/brudi/cmd"
	"github.com/mittwald/brudi/internal"
)

func init() {
	internal.InitLogger()
}

func main() {
	err := cmd.Execute()
	if err != nil && err != cobra.ErrSubCommandRequired {
		panic(err)
	}
}
