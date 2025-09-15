package pgrestore

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
			PGRestore:      false,
		},
	}

	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) RestoreBackup(ctx context.Context) error {
	// Check if SourceFile is a directory (i.e., a pg_dump directory format)
	src := b.cfg.Options.SourceFile
	info, statErr := os.Stat(src)
	if statErr != nil {
		return statErr
	}

	fileName := src

	if !info.IsDir() {
		unzippedFileName, err := cli.CheckAndGunzipFile(src)
		if err != nil {
			return err
		}

		fileName = unzippedFileName
	}

	args := append(cli.StructToCLI(b.cfg.Options.Flags), b.cfg.Options.AdditionalArgs...)
	args = append(args, fileName)
	cmd := cli.CommandType{
		Binary: binary,
		Args:   args,
	}
	var out []byte
	var err error
	out, err = cli.Run(ctx, cmd)
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
