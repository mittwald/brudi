package fsbackup

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

type Options struct {
	Path string `validate:"min=1"`
}

type ConfigBasedBackend struct {
	cfg *Config
}

func NewConfigBasedBackend() (*ConfigBasedBackend, error) {
	config := &Config{
		Options: &Options{},
	}

	if err := config.InitFromViper(); err != nil {
		return nil, err
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) CreateBackup(_ context.Context) error {
	path := b.cfg.Options.Path

	info, err := os.Stat(path)
	if err != nil {
		return errors.WithStack(err)
	}

	if !info.IsDir() {
		return errors.WithStack(fmt.Errorf("configured path %s is not a directory", path))
	}

	return nil
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	return b.cfg.Options.Path
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.HostName
}

func (b *ConfigBasedBackend) CleanUp() error {
	return nil
}
