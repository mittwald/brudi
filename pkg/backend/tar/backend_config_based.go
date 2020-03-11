package tar

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/cli"
)

type ConfigBasedBackend struct {
	config Config
}

func NewConfigBasedBackend() (*ConfigBasedBackend, error) {
	backend := &ConfigBasedBackend{
		config: Config{
			Options: &Options{
				Flags: &Flags{},
				Paths: []string{},
			},
		},
	}

	err := backend.config.InitFromViper()
	if err != nil {
		return nil, err
	}

	return backend, nil
}

func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) error {
	cmd := cli.CommandType{
		Binary: binary,
		Args:   cli.StructToCLI(b.config.Options),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%+v - %+v", err, out))
	}

	return nil
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	return b.config.Options.Flags.File
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.config.HostName
}
