package cmd

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/tarrestore"

	"github.com/spf13/cobra"
)

var (
	tarRestoreCmd = &cobra.Command{
		Use:   "tarrestore",
		Short: "Restores a tar archive of your desired paths",
		Long:  "Restores given paths from a tar backup",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoRestoreForKind(ctx, tarrestore.Kind, cleanup, useRestic)
			if err != nil {
				logrus.WithError(err).Error("Failed to restore database")
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(tarRestoreCmd)
}
