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
	// obtain backup path for restic
	conf.Restore.Flags.Path = backupPaths[0]
	return &Client{
		Logger: resticLogger,
		Config: conf,
	}, nil
}

func (c *Client) DoResticBackup(ctx context.Context) error {
	c.Logger.Info("running 'restic backup'")

	_, err := initBackup(ctx, c.Config.Global)
	if errors.Is(err, ErrRepoAlreadyInitialized) {
		c.Logger.Info("restic repo is already initialized")
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("error while initializing restic repository: %s", err.Error()))
	} else {
		c.Logger.Info("restic repo initialized successfully")
	}

	var out []byte
	_, out, err = CreateBackup(ctx, c.Config.Global, c.Config.Backup, true)
	if err != nil {
		return errors.WithStack(fmt.Errorf("error while running restic backup: %s - %s", err.Error(), out))
	}

	c.Logger.Info("successfully saved restic stuff")

	return nil
}

func (c *Client) DoResticRestore(ctx context.Context, backupPath string) error {
	c.Logger.Info("running 'restic restore'")
	out, err := RestoreBackup(ctx, c.Config.Global, c.Config.Restore, false)
	if err != nil {
		return errors.WithStack(fmt.Errorf("error while running restic restore: %s - %s", err.Error(), out))
	}
	return nil
}

func (c *Client) DoResticForget(ctx context.Context) error {
	c.Logger.Info("running 'restic forget'")

	removedSnapshots, output, err := Forget(ctx, c.Config.Global, c.Config.Forget)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%s - %s", err.Error(), output))
	}

	c.Logger.WithFields(
		log.Fields{
			"snapshotsRemoved": removedSnapshots,
		},
	).Info("successfully forgot restic snapshots")

	return nil
}

func (c *Client) DoResticPrune(ctx context.Context) error {
	c.Logger.Info("running 'restic prune'")

	output, err := Prune(ctx)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%s - %s", err.Error(), output))
	}

	// as of now (16.06.2023) there is no JSON-output for prune
	// therefore we can just print the output as it is
	c.Logger.WithFields(
		log.Fields{
			"output": output,
		},
	).Info("successfully pruned restic snapshots")

	return nil
}
