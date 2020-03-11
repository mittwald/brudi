package mongodump

import (
	"context"

	"github.com/mongodb/mongo-tools-common/options"
	"github.com/mongodb/mongo-tools/mongodump"
	"github.com/pkg/errors"
)

type ConfigBasedBackend struct {
	dump *mongodump.MongoDump
	cfg  *Config
}

func NewConfigBasedBackend() (*ConfigBasedBackend, error) {
	config := &Config{
		ToolOptions: &options.ToolOptions{
			General:    &options.General{},
			Verbosity:  &options.Verbosity{},
			Connection: &options.Connection{},
			SSL:        &options.SSL{},
			Auth:       &options.Auth{},
			Namespace:  &options.Namespace{},
			Kerberos:   &options.Kerberos{},
			URI:        &options.URI{},
		},
		InputOptions:  &mongodump.InputOptions{},
		OutputOptions: &mongodump.OutputOptions{},
	}
	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}

	return newConfigBasedBackendFromConfig(config)
}

func newConfigBasedBackendFromConfig(cfg *Config) (*ConfigBasedBackend, error) {
	return &ConfigBasedBackend{
		cfg: cfg,
		dump: &mongodump.MongoDump{
			ToolOptions:       cfg.ToolOptions,
			InputOptions:      cfg.InputOptions,
			OutputOptions:     cfg.OutputOptions,
			SkipUsersAndRoles: !cfg.OutputOptions.DumpDBUsersAndRoles,
			ProgressManager:   nil,
			SessionProvider:   nil,
			OutputWriter:      nil,
		},
	}, nil
}

func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) error {
	err := b.dump.Init()
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(b.dump.Dump())
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	out := b.dump.OutputOptions.Out
	if len(b.dump.OutputOptions.Archive) > 0 {
		out = b.dump.OutputOptions.Archive
	}

	return out
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.ToolOptions.Connection.Host
}
