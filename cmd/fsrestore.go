package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/fsrestore"

	"github.com/spf13/cobra"
)

var (
	fsrestoreCmd = &cobra.Command{
		Use:   "fsrestore",
		Short: "Restores directories directly from restic",
		Long:  "Restores directories from restic snapshots without additional processing.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := source.DoRestoreForKind(ctx, fsrestore.Kind, cleanup, useRestic); err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(fsrestoreCmd)
}
