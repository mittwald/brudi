package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/backend"

	"github.com/spf13/cobra"

	"github.com/mittwald/brudi/pkg/backend/mongodump"
)

var (
	mongoDumpCmd = &cobra.Command{
		Use:   "mongodump",
		Short: "Creates a mongodump of your desired server",
		Long:  "Backups a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := backend.DoBackupForKind(ctx, mongodump.Kind, cleanup, useRestic)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(mongoDumpCmd)
}
