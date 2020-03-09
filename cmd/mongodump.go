package cmd

import (
	"context"
	"github.com/mittwald/brudi/pkg/backend/mongo"
	"github.com/mittwald/brudi/pkg/cli/restic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	mongoDumpCmd = &cobra.Command{
		Use:   "mongodump",
		Short: "Creates a mongodump of your desired database",
		Long:  "Backups a given database with given arguments",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logMongoKind := log.WithFields(
				log.Fields{
					"kind": mongo.Kind,
					"task": "mongodump",
				})

			backend, err := mongo.NewBackend()
			if err != nil {
				logMongoKind.WithError(err).Fatal("failed while creating mongodb backend")
			}

			err = backend.CreateBackup()
			if err != nil {
				logMongoKind.WithError(err).Fatal("failed while creating backup")
			}

			if cleanup {
				defer func() {
					err = os.RemoveAll(backend.GetBackupPath())
					logMongoKindPath := logMongoKind.WithFields(
						log.Fields{
							"path": backend.GetBackupPath(),
							"cmd":  "cleanup",
						})
					if err = os.RemoveAll(backend.GetBackupPath()); err != nil {
						logMongoKindPath.WithError(err).Warn("failed to cleanup backup")
					} else {
						logMongoKindPath.Info("successfully cleaned up backup")
					}
				}()
			}

			logMongoKind.Info("finished backing up database")

			if !useRestic {
				return
			}

			logMongoKindRestic := logMongoKind.WithField("cmd", "restic")

			logMongoKindRestic.Info("running restic backup")

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
			if err == restic.RepoAlreadyInitialized {
				logMongoKindRestic.Info("restic repo is already initialized")
			} else if err != nil {
				logMongoKindRestic.WithError(err).Fatal("error while initializing restic repository")
			} else {
				logMongoKindRestic.Info("restic repo initialized successfully")
			}

			_, _, err = restic.CreateBackup(ctx, resticBackupOptions, true)
			if err != nil {
				logMongoKindRestic.WithError(err).Fatal("error during restic backup")
			}

			logMongoKindRestic.Info("successfully saved restic stuff")
		},
	}
)

func init() {
	rootCmd.AddCommand(mongoDumpCmd)
}
