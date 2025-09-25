package fsrestore

import (
	"os"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	// Kind identifies the fsrestore backend in configuration.
	Kind = "fsrestore"
)

type Config struct {
	Options  *Options
	HostName string `validate:"min=1"`
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(Kind, c)
	if err != nil {
		return errors.WithStack(err)
	}

	if c.HostName == "" {
		c.HostName, err = os.Hostname()
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return config.Validate(c)
}
