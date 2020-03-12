package mongodump

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "mongodump"
)

type Config struct {
	Options *Flags
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(fmt.Sprintf("%s.%s", Kind, "flags"), c.Options)
	if err != nil {
		return errors.WithStack(err)
	}

	return config.Validate(c, configStructLevelValidation)
}

func configStructLevelValidation(sl validator.StructLevel) {
	c := sl.Current().Interface().(Config)

	if c.Options.Out == "" && c.Options.Archive == "" {
		sl.ReportError(c.Options.Out, "out", "Out", "eitherOutOrArchiveRequired", "")
		sl.ReportError(c.Options.Archive, "archive", "Archive", "eitherOutOrArchiveRequired", "")
	}
}
