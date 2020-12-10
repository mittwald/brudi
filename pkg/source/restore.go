package source

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/restic"

	"github.com/mittwald/brudi/pkg/source/mongorestore"
)

func getGenericRestoreBackendForKind(kind string) (GenericRestore, error) {
	switch kind {
	case mongorestore.Kind:
		return mongorestore.NewConfigBasedBackend()
	default:
		return nil, fmt.Errorf("unsupported kind '%s'", kind)
	}
}

func DoRestoreForKind(ctx context.Context, kind string, cleanup, useRestic, useResticForget bool) error {
	logKind := log.WithFields(
		log.Fields{
			"kind": kind,
		},
	)

	backend, err := getGenericRestoreBackendForKind(kind)
	if err != nil {
		return err
	}

	if !useRestic {
		return nil
	}

	var resticClient *restic.Client
	resticClient, err = restic.NewResticClient(logKind, backend.GetHostname(), backend.GetBackupPath())
	if err != nil {
		return err
	}

	err = resticClient.DoResticRestore(ctx)
	if err != nil {
		return err
	}

	err = backend.RestoreBackup(ctx)
	if err != nil {
		return err
	}

	if cleanup {
		defer func() {
			cleanupLogger := logKind.WithFields(
				log.Fields{
					"path": backend.GetBackupPath(),
					"cmd":  "cleanup",
				},
			)
			if err := backend.CleanUp(); err != nil {
				cleanupLogger.WithError(err).Warn("failed to cleanup backup")
			} else {
				cleanupLogger.Info("successfully cleaned up backup")
			}
		}()
	}

	logKind.Info("finished restoring")

	if !useResticForget {
		return nil
	}

	return resticClient.DoResticForget(ctx)
}
