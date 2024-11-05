package cmd

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/psql"

	"github.com/spf13/cobra"
)

var (
	psqlCmd = &cobra.Command{
		Use:   "psql",
		Short: "restores from plain text pgdump",
		Long:  "Restores a given database with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoRestoreForKind(ctx, psql.Kind, cleanup, useRestic)
			if err != nil {
				logrus.WithError(err).Error("Failed to restore database")
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(psqlCmd)
}
