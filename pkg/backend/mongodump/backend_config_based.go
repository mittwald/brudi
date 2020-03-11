package mongodump

import (
	"context"
	"fmt"

	"github.com/mittwald/brudi/pkg/cli"

	"github.com/pkg/errors"
)

type ConfigBasedBackend struct {
	cfg *Config
}

func NewConfigBasedBackend() (*ConfigBasedBackend, error) {
	config := &Config{
		Options: &Flags{},
	}
	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}

	return &ConfigBasedBackend{cfg: config}, nil
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
	if len(b.cfg.Options.Archive) > 0 {
		return b.cfg.Options.Archive
	}

	return b.cfg.Options.Out
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.Options.Host
}
