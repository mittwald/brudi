package mongorestore

import (
	"context"
	"fmt"
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

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) RestoreBackup(ctx context.Context) error {
	cmd := cli.CommandType{
		Binary: binary,
		Args:   cli.StructToCLI(b.cfg.Options),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%+v - %s", err, out))
	}

	return nil
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
