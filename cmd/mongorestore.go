package cmd

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/spf13/cobra"

	"github.com/mittwald/brudi/pkg/source/mongorestore"
)

var (
	mongoRestoreCmd = &cobra.Command{
		Use:   "mongorestore",
		Short: "restores from mongodump ",
		Long:  "Restores a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoRestoreForKind(ctx, mongorestore.Kind, cleanup, useRestic, useResticForget)
			if err != nil {
				logrus.WithError(err).Error("Failed to restore database")
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(mongoRestoreCmd)
}
