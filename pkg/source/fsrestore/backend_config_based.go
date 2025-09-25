package fsrestore

import "context"

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

func (b *ConfigBasedBackend) RestoreBackup(context.Context) error {
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
