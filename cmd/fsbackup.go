package cmd

import (
	"context"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/fsbackup"

	"github.com/spf13/cobra"
)

var (
	fsbackupCmd = &cobra.Command{
		Use:   "fsbackup",
		Short: "Backs up directories directly with restic",
		Long:  "Backs up configured directories using restic without creating intermediate archives.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := source.DoBackupForKind(ctx, fsbackup.Kind, cleanup, useRestic, useResticForget, useResticPrune); err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(fsbackupCmd)
}
