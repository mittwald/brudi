package redisdump

import (
	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "redisdump"
)

type Config struct {
	Options *Options
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(Kind, c)
	if err != nil {
		return errors.WithStack(err)
	}

	return config.Validate(c)
}
