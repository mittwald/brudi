package cmd

import (
	"context"
	"github.com/mittwald/brudi/pkg/source/directoryrestore"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/spf13/cobra"
)

var (
	directoryRestoreCmd = &cobra.Command{
		Use:   "directory",
		Short: "Restores given directory",
		Long:  "Restores given directory",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoRestoreForKind(ctx, directoryrestore.Kind, cleanup, useRestic)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(directoryRestoreCmd)
}
