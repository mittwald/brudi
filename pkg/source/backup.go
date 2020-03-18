package source

import (
	"context"
	"fmt"

	"github.com/mittwald/brudi/pkg/source/pgdump"

	"github.com/mittwald/brudi/pkg/source/tar"

	log "github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/source/mongodump"
	"github.com/mittwald/brudi/pkg/source/mysqldump"
)

func getGenericBackendForKind(kind string) (Generic, error) {
	switch kind {
	case pgdump.Kind:
		return pgdump.NewConfigBasedBackend()
	case mongodump.Kind:
		return mongodump.NewConfigBasedBackend()
	case mysqldump.Kind:
		return mysqldump.NewConfigBasedBackend()
	case tar.Kind:
		return tar.NewConfigBasedBackend()
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
			if err := backend.CleanUp(); err != nil {
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
