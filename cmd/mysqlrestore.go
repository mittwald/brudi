package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/mysqlrestore"
)

var (
	mysqlRestoreCmd = &cobra.Command{
		Use:   "mysqlrestore",
		Short: "restores from mysqldump ",
		Long:  "Restores a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := source.DoRestoreForKind(ctx, mysqlrestore.Kind, cleanup, useRestic)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(mysqlRestoreCmd)
}
