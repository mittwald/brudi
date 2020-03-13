package tar

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
			Flags: &Flags{},
			Paths: []string{},
		},
	}

	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) error {
	var args []string
	args = append(args, cli.StructToCLI(b.cfg.Options.Flags)...)
	args = append(args, b.cfg.Options.AdditionalArgs...)

	cmd := cli.CommandType{
		Binary: binary,
		Args:   args,
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

func (b *ConfigBasedBackend) CleanUp() error {
	return os.Remove(b.GetBackupPath())
}
