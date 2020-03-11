package tar

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/cli"
)

type ConfigBasedBackend struct {
	cfg *Config
}

func NewConfigBasedBackend() (*ConfigBasedBackend, error) {
	backend := &ConfigBasedBackend{
		cfg: &Config{
			Options: &Options{
				Flags: &Flags{},
				Paths: []string{},
			},
		},
	}

	err := backend.cfg.InitFromViper()
	if err != nil {
		return nil, err
	}

	return backend, nil
}

func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) error {
	cmd := cli.CommandType{
		Binary: binary,
		Args:   cli.StructToCLI(b.cfg.Options),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%+v - %+v", err, out))
	}

	return nil
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	return b.cfg.Options.Flags.File
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.HostName
}
