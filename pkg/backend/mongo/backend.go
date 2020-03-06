package mongo

import (
	"github.com/mongodb/mongo-tools-common/options"
	"github.com/mongodb/mongo-tools/mongodump"
	"github.com/pkg/errors"
)

type Backend struct {
	dump *mongodump.MongoDump
	cfg  *Config
}

func NewBackend() (*Backend, error) {
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

	return newBackendFromConfig(config)
}

func newBackendFromConfig(cfg *Config) (*Backend, error) {
	return &Backend{
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

func (b *Backend) CreateBackup() error {
	err := b.dump.Init()
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(b.dump.Dump())
}

func (b *Backend) GetBackupPath() string {
	out := b.dump.OutputOptions.Out
	if len(b.dump.OutputOptions.Archive) > 0 {
		out = b.dump.OutputOptions.Archive
	}

	return out
}

func (b *Backend) GetHostname() string {
	return b.cfg.ToolOptions.Connection.Host
}
