package backend

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/backend/mongodump"
	"github.com/mittwald/brudi/pkg/backend/mysqldump"
)

func getGenericBackendForKind(kind string) (Generic, error) {
	switch kind {
	case mongodump.Kind:
		return mongodump.NewConfigBasedBackend()
	case mysqldump.Kind:
		return mysqldump.NewConfigBasedBackend()
	default:
		return nil, fmt.Errorf("unsupported kind '%s'", kind)
	}
}

func DoBackupForKind(ctx context.Context, kind string, cleanup, useRestic bool) error {
	logKind := log.WithFields(
		log.Fields{
			"kind": kind,
		},
	)

	backend, err := getGenericBackendForKind(kind)
	if err != nil {
		return err
	}

	err = backend.CreateBackup(ctx)
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
			if err := os.RemoveAll(backend.GetBackupPath()); err != nil {
				cleanupLogger.WithError(err).Warn("failed to cleanup backup")
			} else {
				cleanupLogger.Info("successfully cleaned up backup")
			}
		}()
	}

	logKind.Info("finished backing up")

	if !useRestic {
		return nil
	}

	resticWrapper := NewResticWrapper(logKind, backend)

	err = resticWrapper.DoRestic(ctx)
	if err != nil {
		return err
	}

	return nil
}