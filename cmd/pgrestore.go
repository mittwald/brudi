package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/spf13/cobra"

	"github.com/mittwald/brudi/pkg/source/pgrestore"
)

var (
	pgRestoreCmd = &cobra.Command{
		Use:   "pgrestore",
		Short: "restores from pgdump",
		Long:  "Restores a given database with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoRestoreForKind(ctx, pgrestore.Kind, cleanup, useRestic)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(pgRestoreCmd)
}
