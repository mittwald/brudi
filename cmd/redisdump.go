package cmd

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/source/redisdump"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/spf13/cobra"
)

var (
	redisDumpCmd = &cobra.Command{
		Use:   "redisdump",
		Short: "Creates an rdb dump of your desired server",
		Long:  "Backups a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoBackupForKind(ctx, redisdump.Kind, cleanup, useRestic, useResticForget)
			if err != nil {
				logrus.WithError(err).Error("Failed to backup database")
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(redisDumpCmd)
}
