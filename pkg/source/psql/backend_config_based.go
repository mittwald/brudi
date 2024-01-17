package psql

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"

	"github.com/mittwald/brudi/pkg/cli"

	"github.com/pkg/errors"
)

//var _ source.GenericRestore = &ConfigBasedBackend{}

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
	if viper.GetBool(cli.DoStdinBackupKey) {
		config.Options.Flags.Output = ""
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) RestoreBackup(ctx context.Context) error {
	fileName, err := cli.CheckAndGunzipFile(b.cfg.Options.SourceFile)
	if err != nil {
		return err
	}
	if b.cfg.Options.Flags.Command == "" {
		b.cfg.Options.Flags.Command = fmt.Sprintf("\\i %s", fileName)
	}
	args := append(cli.StructToCLI(b.cfg.Options.Flags), b.cfg.Options.AdditionalArgs...)
	cmd := cli.CommandType{
		Binary: binary,
		Args:   args,
	}
	var out []byte
	out, err = cli.Run(ctx, &cmd, false)
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
