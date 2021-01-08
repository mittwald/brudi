package mysqlrestore

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
		&Options{
			Flags:          &Flags{},
			AdditionalArgs: []string{},
			SourceFile:     "",
		},
	}
	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) RestoreBackup(ctx context.Context) error {
	fileName, err := cli.CheckAndGunzipFile(b.cfg.Options.SourceFile)
	if b.cfg.Options.Flags.Execute == "" {
		b.cfg.Options.Flags.Execute = fmt.Sprintf("source %s", fileName)
	}
	args := append(cli.StructToCLI(b.cfg.Options.Flags), b.cfg.Options.AdditionalArgs...)
	cmd := cli.CommandType{
		Binary: binary,
		Args:   args,
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%+v - %s", err, out))
	}

	return nil
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	return b.cfg.Options.SourceFile
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.Options.Flags.Host
}

func (b *ConfigBasedBackend) CleanUp() error {
	return os.Remove(b.GetBackupPath())
}
