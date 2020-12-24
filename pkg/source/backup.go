package source

import (
	"context"
	"fmt"

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

func DoBackupForKind(ctx context.Context, kind string, cleanup, useRestic, useResticForget, listResticSnapshots,
	resticCheck, resticPrune, rebuildIndex bool) error {
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

	var resticClient *restic.Client
	resticClient, err = restic.NewResticClient(logKind, backend.GetHostname(), backend.GetBackupPath())
	if err != nil {
		return err
	}

	err = resticClient.DoResticBackup(ctx)
	if err != nil {
		return err
	}

	if useResticForget {
		err = resticClient.DoResticForget(ctx)
		if err != nil {
			return err
		}
	}
	if listResticSnapshots {
		err = resticClient.ListSnapshots(ctx)
		if err != nil {
			return err
		}
	}

	if resticPrune {
		err = resticClient.PruneRepo(ctx)
		if err != nil {
			return err
		}
	}

	if rebuildIndex {
		err = resticClient.RebuildIndex(ctx)
		if err != nil {
			return err
		}
	}

	if !resticCheck {
		return nil
	}

	return resticClient.DoResticCheck(ctx)

}
