package cmd

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/spf13/cobra"

	"github.com/mittwald/brudi/pkg/source/mongodump"
)

var (
	mongoDumpCmd = &cobra.Command{
		Use:   "mongodump",
		Short: "Creates a mongodump of your desired server",
		Long:  "Backups a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoBackupForKind(ctx, mongodump.Kind, cleanup, useRestic, useResticForget)
			if err != nil {
				logrus.WithError(err).Error("Failed to backup database")
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(mongoDumpCmd)
}
