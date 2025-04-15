package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/mittwald/brudi/pkg/source/directory"

	"github.com/spf13/cobra"
)

var (
	directoryCmd = &cobra.Command{
		Use:   "directory",
		Short: "Backups given directory",
		Long:  "Backups given directory",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoBackupForKind(ctx, directory.Kind, cleanup, useRestic, useResticForget, useResticPrune)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(directoryCmd)
}
