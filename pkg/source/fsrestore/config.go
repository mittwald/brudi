package fsrestore

import (
	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	// Kind identifies the fsrestore backend in configuration.
	Kind = "fsrestore"
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
