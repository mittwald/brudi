package tarrestore

import (
	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "tarrestore"
)

type Config struct {
	Options  *Options
	HostName string
}

func (c *Config) InitFromViper() error {
	if err := config.InitializeStructFromViper(Kind, c); err != nil {
		return errors.WithStack(err)
	}

	return config.Validate(c)
}
