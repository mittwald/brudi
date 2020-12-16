package source

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/restic"
	"github.com/mittwald/brudi/pkg/source/mongorestore"
	"github.com/mittwald/brudi/pkg/source/mysqlrestore"
	"github.com/mittwald/brudi/pkg/source/pgrestore"
	"github.com/mittwald/brudi/pkg/source/redisrestore"
)

func getGenericRestoreBackendForKind(kind string) (GenericRestore, error) {
	switch kind {
	case mongorestore.Kind:
		return mongorestore.NewConfigBasedBackend()
	case mysqlrestore.Kind:
		return mysqlrestore.NewConfigBasedBackend()
	case pgrestore.Kind:
		return pgrestore.NewConfigBasedBackend()
	case redisrestore.Kind:
		return redisrestore.NewConfigBasedBackend()
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

	if useRestic {
		var resticClient *restic.Client
		resticClient, err = restic.NewResticClient(logKind, backend.GetHostname(), backend.GetBackupPath())
		if err != nil {
			return err
		}

		err = resticClient.DoResticRestore(ctx)
		if err != nil {
			return err
		}

		if useResticForget {
			err = resticClient.DoResticForget(ctx)
			if err != nil {
				return err
			}
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
			if err := backend.CleanUp(); err != nil {
				cleanupLogger.WithError(err).Warn("failed to cleanup backup")
			} else {
				cleanupLogger.Info("successfully cleaned up backup")
			}
		}()
	}

	logKind.Info("finished restoring")

	return nil
}
