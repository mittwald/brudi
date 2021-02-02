package main

import (
	"errors"

	"github.com/mittwald/brudi/cmd"
	"github.com/mittwald/brudi/internal"
	"github.com/spf13/cobra"
)

func init() {
	internal.InitLogger()
}

func main() {
	err := cmd.Execute()
	if err != nil && !errors.Is(err, cobra.ErrSubCommandRequired) {
		panic(err)
	}
}
