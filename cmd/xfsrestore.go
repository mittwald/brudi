package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/xfsrestore"

	"github.com/spf13/cobra"
)

var (
	xfsRestoreCmd = &cobra.Command{
		Use:   "xfsrestore",
		Short: "Restores a xfsdump of A file system",
		Long:  "Restores a given filesystem with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoRestoreForKind(ctx, xfsrestore.Kind, cleanup, useRestic, useResticForget)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(xfsRestoreCmd)
}
