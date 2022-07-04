package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/spf13/cobra"

	"github.com/mittwald/brudi/pkg/source/xfsdump"
)

var (
	xfsDumpCmd = &cobra.Command{
		Use:   "xfsdump",
		Short: "Creates a xfsdump of your desired file system",
		Long:  "Backups a given filesystem with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoBackupForKind(ctx, xfsdump.Kind, cleanup, useRestic, useResticForget)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(xfsDumpCmd)
}
