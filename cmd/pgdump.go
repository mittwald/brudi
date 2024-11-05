package cmd

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/source/pgdump"

	"github.com/spf13/cobra"

	"github.com/mittwald/brudi/pkg/source"
)

var (
	pgDumpCmd = &cobra.Command{
		Use:   "pgdump",
		Short: "Creates a pg_dump of your desired postgresql-server",
		Long:  "Backups a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoBackupForKind(ctx, pgdump.Kind, cleanup, useRestic, useResticForget)
			if err != nil {
				logrus.WithError(err).Error("Failed to backup database")
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(pgDumpCmd)
}
