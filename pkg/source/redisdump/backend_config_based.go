package redisdump

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/cli"
)

//var _ source.Generic = &ConfigBasedBackend{}

type ConfigBasedBackend struct {
	cfg *Config
}

func NewConfigBasedBackend() (*ConfigBasedBackend, error) {
	if viper.GetBool(cli.DoStdinBackupKey) {
		//config.Options.Flags.Pipe = true
		return nil, errors.Errorf("can't do a backup to STDOUT with redisdump but %s is set", cli.DoStdinBackupKey)
	}

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
func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) (*cli.CommandType, error) {
	gzip := false
	// create temporary, unzipped backup first, thus trim '.gz' extension
	if strings.HasSuffix(b.cfg.Options.Flags.Rdb, cli.GzipSuffix) {
		b.cfg.Options.Flags.Rdb = strings.TrimSuffix(b.cfg.Options.Flags.Rdb, cli.GzipSuffix)
		gzip = true
	}
	cmd := b.GetBackupCommand()

	out, err := cli.Run(ctx, &cmd, false)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("%+v - %s", err, out))
	}

	// zip backup, update flag with the name returned by GzipFile for correct handover to restic
	if gzip {
		b.cfg.Options.Flags.Rdb, err = cli.GzipFile(b.cfg.Options.Flags.Rdb)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (b *ConfigBasedBackend) GetBackupCommand() cli.CommandType {
	return cli.CommandType{
		Binary: binary,
		Args:   cli.StructToCLI(b.cfg.Options),
	}
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
