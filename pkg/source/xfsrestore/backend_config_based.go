package xfsrestore

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/cli"
)

type ConfigBasedBackend struct {
	cfg *Config
}

func NewConfigBasedBackend() (*ConfigBasedBackend, error) {
	config := &Config{
		Options: &Options{
			Flags:          &Flags{},
			AdditionalArgs: []string{},
			DestFS:         "",
		},
	}

	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) RestoreBackup(ctx context.Context) error {
	cmd := cli.CommandType{
		Binary: binary,
		Args:   append(cli.StructToCLI(b.cfg.Options), b.cfg.Options.DestFS),
	}

	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%+v - %s", err, out))
	}

	return nil
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	return b.cfg.Options.Flags.Source
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.HostName
}

func (b *ConfigBasedBackend) CleanUp() error {
	return os.Remove(b.GetBackupPath())
}
