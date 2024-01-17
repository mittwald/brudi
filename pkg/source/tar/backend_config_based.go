package tar

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
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
			Paths:          []string{},
			AdditionalArgs: []string{},
		},
	}

	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}
	if viper.GetBool(cli.DoStdinBackupKey) {
		config.Options.Flags.File = "-"
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) error {
	cmd := b.GetBackupCommand()
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%+v - %s", err, out))
	}

	return nil
}

func (b *ConfigBasedBackend) GetBackupCommand() cli.CommandType {
	return cli.CommandType{
		Binary: binary,
		Args:   cli.StructToCLI(b.cfg.Options),
	}
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	return b.cfg.Options.Flags.File
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.HostName
}

func (b *ConfigBasedBackend) CleanUp() error {
	return os.Remove(b.GetBackupPath())
}
