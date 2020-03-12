package backend

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/restic"

	log "github.com/sirupsen/logrus"
)

type Wrapper struct {
	Logger        *log.Entry
	BackupOptions *restic.BackupOptions
	Backend       Generic
}

func NewResticWrapper(logger *log.Entry, be Generic) *Wrapper {
	options := &restic.BackupOptions{
		Flags: &restic.BackupFlags{
			Host: be.GetHostname(),
		},
		Paths: []string{
			be.GetBackupPath(),
		},
	}

	resticLogger := logger.WithField("cmd", "restic")

	resticLogger.Info("running restic backup")

	return &Wrapper{
		Logger:        resticLogger,
		BackupOptions: options,
		Backend:       be,
	}
}

func (w *Wrapper) DoRestic(ctx context.Context) error {
	_ = os.Setenv("RESTIC_HOST", w.Backend.GetHostname())

	_, err := restic.Init()
	if err == restic.ErrRepoAlreadyInitialized {
		w.Logger.Info("restic repo is already initialized")
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("error while initializing restic repository: %s", err.Error()))
	} else {
		w.Logger.Info("restic repo initialized successfully")
	}

	_, _, err = restic.CreateBackup(ctx, w.BackupOptions, true)
	if err != nil {
		return errors.WithStack(fmt.Errorf("error while while running restic backup: %s", err.Error()))
	}

	w.Logger.Info("successfully saved restic stuff")

	return nil
}
