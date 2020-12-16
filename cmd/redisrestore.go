package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/redisrestore"

	"github.com/spf13/cobra"
)

var (
	redisRestoreCmd = &cobra.Command{
		Use:   "redisrestore",
		Short: "restores from rdb file",
		Long:  "Restores a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoRestoreForKind(ctx, redisrestore.Kind, cleanup, useRestic, useResticForget)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(redisRestoreCmd)
}
