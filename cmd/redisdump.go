package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source/redisdump"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/spf13/cobra"
)

var (
	redisDumpCmd = &cobra.Command{
		Use:   "redisdump",
		Short: "Creates a redisdump of your desired server",
		Long:  "Backups a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoBackupForKind(ctx, redisdump.Kind, cleanup, useRestic, useResticForget)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(redisDumpCmd)
}
