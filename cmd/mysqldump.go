package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/backend"
	"github.com/mittwald/brudi/pkg/backend/mysqldump"

	"github.com/spf13/cobra"
)

var (
	mysqlDumpCmd = &cobra.Command{
		Use:   "mysqldump",
		Short: "Creates a mysqldump of your desired server",
		Long:  "Backups a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := backend.DoBackupForKind(ctx, mysqldump.Kind, cleanup, useRestic)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(mysqlDumpCmd)
}
