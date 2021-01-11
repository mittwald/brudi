package redisdump

import (
	"context"
	"fmt"
	"os"
	"strings"

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
			Command:        "bgsave",
		},
	}

	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

// Do a bgsave of the given redis instance
func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) error {
	var gzip bool
	if strings.HasSuffix(b.cfg.Options.Flags.Rdb, cli.GzipSuffix) {
		b.cfg.Options.Flags.Rdb = strings.TrimRight(b.cfg.Options.Flags.Rdb, cli.GzipSuffix)
		gzip = true
	}
	cmd := cli.CommandType{
		Binary: binary,
		Args:   cli.StructToCLI(b.cfg.Options),
	}

	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%+v - %s", err, out))
	}

	if gzip {
		b.cfg.Options.Flags.Rdb, err = cli.GzipFile(b.cfg.Options.Flags.Rdb)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	return b.cfg.Options.Flags.Rdb
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.Options.Flags.Host
}

func (b *ConfigBasedBackend) CleanUp() error {
	return os.Remove(b.GetBackupPath())
}
