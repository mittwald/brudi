package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/mysqldump"

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

			err := source.DoBackupForKind(ctx, mysqldump.Kind, cleanup, useRestic, useResticForget, gzip)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(mysqlDumpCmd)
}
