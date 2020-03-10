package cmd

import (
	"context"
	"github.com/mittwald/brudi/pkg/backend/mysqldump"
	"github.com/mittwald/brudi/pkg/cli/restic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	mysqlDumpCmd = &cobra.Command{
		Use:   "mysqldump",
		Short: "Creates a mysqldump of your desired server",
		Long:  "Backups a given database server with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logMysqlKind := log.WithFields(
				log.Fields{
					"kind": mysqldump.Kind,
					"task": "mysqldump",
				},
			)

			backend, err := mysqldump.NewBackend()
			if err != nil {
				logMysqlKind.WithError(err).Fatal("failed while creating backend")
			}

			err = backend.CreateBackup(ctx)
			if err != nil {
				logMysqlKind.WithError(err).Fatal("unable to create dump")
			}

			if cleanup {
				defer func() {
					err = os.RemoveAll(backend.GetBackupPath())
					logMysqlKindPath := logMysqlKind.WithFields(
						log.Fields{
							"path": backend.GetBackupPath(),
							"cmd":  "cleanup",
						})
					if err = os.RemoveAll(backend.GetBackupPath()); err != nil {
						logMysqlKindPath.WithError(err).Warn("failed to cleanup backup")
					} else {
						logMysqlKindPath.Info("successfully cleaned up backup")
					}
				}()
			}

			logMysqlKind.Info("finished backing up database")

			if !useRestic {
				return
			}

			logMysqlKindRestic := logMysqlKind.WithField("cmd", "restic")

			logMysqlKindRestic.Info("running restic backup")

			resticBackupOptions := &restic.BackupOptions{
				Flags: &restic.BackupFlags{
					Host: backend.GetHostname(),
				},
				Paths: []string{
					backend.GetBackupPath(),
				},
			}

			_ = os.Setenv("RESTIC_HOST", backend.GetHostname())

			_, err = restic.Init()
			if err == restic.ErrRepoAlreadyInitialized {
				logMysqlKindRestic.Info("restic repo is already initialized")
			} else if err != nil {
				logMysqlKindRestic.WithError(err).Fatal("error while initializing restic repository")
			} else {
				logMysqlKindRestic.Info("restic repo initialized successfully")
			}

			_, _, err = restic.CreateBackup(ctx, resticBackupOptions, true)
			if err != nil {
				logMysqlKindRestic.WithError(err).Fatal("error during restic backup")
			}

			logMysqlKindRestic.Info("successfully saved restic stuff")
		},
	}
)

func init() {
	rootCmd.AddCommand(mysqlDumpCmd)
}
