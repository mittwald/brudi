package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source/tar"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/spf13/cobra"
)

var (
	tarCmd = &cobra.Command{
		Use:   "tar",
		Short: "Creates a tar archive of your desired paths",
		Long:  "Backups given paths by creating a tar backup",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoBackupForKind(ctx, tar.Kind, cleanup, useRestic, useResticForget, gzip)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(tarCmd)
}
