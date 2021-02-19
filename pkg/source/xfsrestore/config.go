package xfsrestore

import (
	"os"

	"github.com/go-playground/validator/v10"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "xfsrestore"
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

	return config.Validate(c, configStructLevelValidation)
}

func configStructLevelValidation(sl validator.StructLevel) {
	c := sl.Current().Interface().(Config)

	if c.Options.Flags.Source == "" {
		sl.ReportError(c.Options.Flags.Source, "source", "Source", "sourceRequired", "")
	}
}
