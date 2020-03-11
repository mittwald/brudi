package cmd

import (
	"context"
	"github.com/mittwald/brudi/pkg/backend/tar"

	"github.com/mittwald/brudi/pkg/backend"

	"github.com/spf13/cobra"
)

var (
	tarCmd = &cobra.Command{
		Use:   "tar",
		Short: "Creates a tar archive of your desired path",
		Long:  "Backups a given path by creating a tar backup",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := backend.DoBackupForKind(ctx, tar.Kind, cleanup, useRestic)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(tarCmd)
}
