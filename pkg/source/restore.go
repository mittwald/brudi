package source

import (
	"context"
	"fmt"
	"github.com/mittwald/brudi/pkg/source/directoryrestore"

	log "github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/restic"
	"github.com/mittwald/brudi/pkg/source/mongorestore"
	"github.com/mittwald/brudi/pkg/source/mysqlrestore"
	"github.com/mittwald/brudi/pkg/source/pgrestore"
	"github.com/mittwald/brudi/pkg/source/psql"
	"github.com/mittwald/brudi/pkg/source/tarrestore"
)

func getGenericRestoreBackendForKind(kind string) (GenericRestore, error) {
	switch kind {
	case mongorestore.Kind:
		return mongorestore.NewConfigBasedBackend()
	case mysqlrestore.Kind:
		return mysqlrestore.NewConfigBasedBackend()
	case pgrestore.Kind:
		return pgrestore.NewConfigBasedBackend()
	case tarrestore.Kind:
		return tarrestore.NewConfigBasedBackend()
	case psql.Kind:
		return psql.NewConfigBasedBackend()
	case directoryrestore.Kind:
		return directoryrestore.NewConfigBasedBackend()
	default:
		return nil, fmt.Errorf("unsupported kind '%s'", kind)
	}
}

func DoRestoreForKind(ctx context.Context, kind string, cleanup, useRestic bool) error {
	logKind := log.WithFields(
		log.Fields{
			"kind": kind,
		},
	)

	backend, err := getGenericRestoreBackendForKind(kind)
	if err != nil {
		return err
	}

	if useRestic { // nolint: nestif
		var resticClient *restic.Client
		resticClient, err = restic.NewResticClient(logKind, backend.GetHostname(), backend.GetBackupPath())
		if err != nil {
			return err
		}

		err = resticClient.DoResticRestore(ctx, backend.GetBackupPath())
		if err != nil {
			return err
		}
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
			if cleanupErr := backend.CleanUp(); cleanupErr != nil {
				cleanupLogger.WithError(cleanupErr).Warn("failed to cleanup backup")
			} else {
				cleanupLogger.Info("successfully cleaned up backup")
			}
		}()
	}

	logKind.Info("finished restoring")

	return nil
}
