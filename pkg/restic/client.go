package restic

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Logger *log.Entry
	Config *Config
}

func NewResticClient(logger *log.Entry, hostname string, backupPaths ...string) (*Client, error) {
	conf := &Config{
		Global: &GlobalOptions{
			Flags: &GlobalFlags{},
		},
		Backup: &BackupOptions{
			Flags: &BackupFlags{},
			Paths: []string{},
		},
		Forget: &ForgetOptions{
			Flags: &ForgetFlags{},
			IDs:   []string{},
		},
		Restore: &RestoreOptions{
			Flags: &RestoreFlags{},
			ID:    "",
		},
	}

	err := conf.InitFromViper()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if (conf.Backup.Flags.Host) == "" {
		conf.Backup.Flags.Host = hostname
	}

	conf.Backup.Paths = append(conf.Backup.Paths, backupPaths...)
	resticLogger := logger.WithField("cmd", "restic")
	conf.Restore.Flags.Path = backupPaths[0]
	return &Client{
		Logger: resticLogger,
		Config: conf,
	}, nil
}

func (c *Client) DoResticBackup(ctx context.Context) error {
	c.Logger.Info("running 'restic backup'")

	_, err := initBackup(ctx, c.Config.Global)
	if err == ErrRepoAlreadyInitialized {
		c.Logger.Info("restic repo is already initialized")
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("error while initializing restic repository: %s", err.Error()))
	} else {
		c.Logger.Info("restic repo initialized successfully")
	}

	_, _, err = CreateBackup(ctx, c.Config.Global, c.Config.Backup, true)
	if err != nil {
		return errors.WithStack(fmt.Errorf("error while while running restic backup: %s", err.Error()))
	}

	c.Logger.Info("successfully saved restic stuff")

	return nil
}

func (c *Client) DoResticRestore(ctx context.Context, backupPath string) error {
	c.Logger.Info("running 'restic restore'")
	_, err := RestoreBackup(ctx, c.Config.Global, c.Config.Restore, false)
	if err != nil {
		return errors.WithStack(fmt.Errorf("error while while running restic restore: %s", err.Error()))
	}
	return nil
}

func (c *Client) DoResticForget(ctx context.Context) error {
	c.Logger.Info("running 'restic forget'")

	removedSnapshots, output, err := Forget(ctx, c.Config.Global, c.Config.Forget)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%s - %s", err.Error(), output))
	}

	c.Logger.WithFields(log.Fields{
		"snapshotsRemoved": removedSnapshots,
	}).Info("successfully forgot restic snapshots")

	return nil
}
