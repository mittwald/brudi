package restic

import (
	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "restic"
)

type Config struct {
	Global *GlobalOptions
	Backup *BackupOptions
	Forget *ForgetOptions
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(Kind, c)
	if err != nil {
		return err
	}

	return config.Validate(c)
}
