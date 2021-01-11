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

// NewResticClient creates a new restic client with the given hostname and backup paths
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
		Snapshots: &SnapshotOptions{
			Flags: &SnapshotFlags{},
			IDs:   []string{},
		},
		Tags: &TagOptions{
			Flags: &TagFlags{},
			IDs:   []string{},
		},
		Check: &CheckFlags{},
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

// DoResticBackup executes initBackup and CreateBackup with the settings from c
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

// DoResticListSnapshots executes ListSnapshots with the settings from c
func (c *Client) DoResticListSnapshots(ctx context.Context) error {
	c.Logger.Info("running 'restic snapshots'")

	output, err := ListSnapshots(ctx, c.Config.Global, c.Config.Snapshots)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%s - %s", err.Error(), output))
	}
	fmt.Println("output of 'restic snapshots':")
	for index := range output {
		fmt.Println(fmt.Sprintf("ID: %s; Time: %s; Host: %s; Tags: %s; Paths: %s",
			output[index].ID, output[index].Time, output[index].Hostname, output[index].Tags, output[index].Paths))
	}
	return nil
}

// DoResticCheck executes Check with the settings from c
func (c *Client) DoResticCheck(ctx context.Context) error {
	c.Logger.Info("running 'restic check'")

	output, err := Check(ctx, c.Config.Global, c.Config.Check)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%s - %s", err.Error(), output))
	}
	fmt.Println(string(output))
	return nil
}

// DoResticPruneRepo executes Prune with the settings from c
func (c *Client) DoResticPruneRepo(ctx context.Context) error {
	c.Logger.Info("running 'restic prune")

	output, err := Prune(ctx, c.Config.Global)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%s - %s", err.Error(), output))
	}
	fmt.Println(string(output))
	return nil
}

// DoResticRebuildIndex executes RebuildIndex with the settings from c
func (c *Client) DoResticRebuildIndex(ctx context.Context) error {
	c.Logger.Info("running 'restic rebuild-index'")

	output, err := RebuildIndex(ctx, c.Config.Global)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%s - %s", err.Error(), output))
	}
	fmt.Println(string(output))
	return nil
}

// DoResticTag executes Tag with the settings from c
func (c *Client) DoResticTag(ctx context.Context) error {
	c.Logger.Info("running 'restic rebuild-index'")

	output, err := Tag(ctx, c.Config.Global, c.Config.Tags)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%s - %s", err.Error(), output))
	}
	fmt.Println(string(output))
	return nil
}
