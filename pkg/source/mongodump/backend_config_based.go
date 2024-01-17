package mongodump

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/internal"
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
		},
	}

	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}
	if viper.GetBool(cli.DoStdinBackupKey) {
		config.Options.Flags.Archive = ""
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
	if b.cfg.Options.Flags.Archive != "" {
		return b.cfg.Options.Flags.Archive
	}

	return b.cfg.Options.Flags.Out
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.Options.Flags.Host
}

func (b *ConfigBasedBackend) CleanUp() error {
	var fileTypes []string

	if b.cfg.Options.Flags.Archive != "" {
		return os.Remove(b.cfg.Options.Flags.Archive)
	} else if b.cfg.Options.Flags.Gzip {
		fileTypes = append(fileTypes, ".bson.gz", ".json.gz")
	} else {
		fileTypes = append(fileTypes, ".bson", ".json")
	}

	return internal.ClearDirectory(b.GetBackupPath(), fileTypes...)
}
