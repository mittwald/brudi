package source

import (
	"context"
	"fmt"
	"github.com/mittwald/brudi/pkg/cli"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/mittwald/brudi/pkg/restic"

	"github.com/mittwald/brudi/pkg/source/pgdump"

	"github.com/mittwald/brudi/pkg/source/tar"

	log "github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/source/mongodump"
	"github.com/mittwald/brudi/pkg/source/mysqldump"
	"github.com/mittwald/brudi/pkg/source/redisdump"
)

func getGenericBackendForKind(kind string) (Generic, error) {
	switch kind {
	case pgdump.Kind:
		return pgdump.NewConfigBasedBackend()
	case mongodump.Kind:
		return mongodump.NewConfigBasedBackend()
	case mysqldump.Kind:
		return mysqldump.NewConfigBasedBackend()
	case redisdump.Kind:
		return redisdump.NewConfigBasedBackend()
	case tar.Kind:
		return tar.NewConfigBasedBackend()
	default:
		return nil, fmt.Errorf("unsupported kind '%s'", kind)
	}
}

func DoBackupForKind(ctx context.Context, kind string, cleanup, useRestic, useResticForget, useResticPrune bool) error {
	if viper.GetBool(cli.DoStdinBackupKey) && !useRestic {
		return errors.New("doStdinBackup is enabled but restic is disabled")
	}
	logKind := log.WithFields(
		log.Fields{
			"kind": kind,
		},
	)

	backend, err := getGenericBackendForKind(kind)
	if err != nil {
		return err
	}

	// TODO: Re-activate when --stdin-command was added to restic
	/*var backupCmd *cli.CommandType = nil
	if viper.GetBool(cli.DoStdinBackupKey) {
		bc := backend.GetBackupCommand()
		backupCmd = &bc
	} else {*/
	backupCmd, err := backend.CreateBackup(ctx)
	if err != nil {
		return err
	}

	if !viper.GetBool(cli.DoStdinBackupKey) {
		if cleanup {
			defer func() {
				cleanupLogger := logKind.WithFields(
					log.Fields{
						"path": backend.GetBackupPath(),
						"cmd":  "cleanup",
					},
				)
				if err = backend.CleanUp(); err != nil {
					cleanupLogger.WithError(err).Warn("failed to cleanup backup")
				} else {
					cleanupLogger.Info("successfully cleaned up backup")
				}
			}()
		}
		logKind.Info("finished backing up")
	}

	if !useRestic {
		return nil
	}

	var resticClient *restic.Client
	resticClient, err = restic.NewResticClient(logKind, backend.GetHostname(), backend.GetBackupPath())
	if err != nil {
		return err
	}

	// as of now (16.06.2023) there is no JSON-output for `restic forget --prune`
	// if we use forget with the `prune`-flag we encounter a parse-error because of invalid json
	// therefore we do not pass the `--prune`-flag to restic but execute `restic prune`
	if resticClient.Config.Forget.Flags.Prune {
		useResticPrune = true
		resticClient.Config.Forget.Flags.Prune = false
	}

	if doBackupErr := resticClient.DoResticBackup(ctx, backupCmd); doBackupErr != nil {
		return doBackupErr
	}

	if useResticForget {
		forgetErr := resticClient.DoResticForget(ctx)
		if forgetErr != nil {
			return forgetErr
		}
	}

	if useResticPrune {
		pruneErr := resticClient.DoResticPrune(ctx)
		if pruneErr != nil {
			return pruneErr
		}
	}

	return nil
}
