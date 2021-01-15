package mongorestore

import (
	"github.com/go-playground/validator/v10"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "mongorestore"
)

type Config struct {
	Options *Options
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(Kind, c)
	if err != nil {
		return errors.WithStack(err)
	}

	return config.Validate(c, configStructLevelValidation)
}

func configStructLevelValidation(sl validator.StructLevel) {
	c := sl.Current().Interface().(Config)

	if c.Options.Flags.Dir == "" && c.Options.Flags.Archive == "" {
		sl.ReportError(c.Options.Flags.Dir, "dir", "Dir", "eitherDirOrArchiveRequired", "")
		sl.ReportError(c.Options.Flags.Archive, "archive", "Archive", "eitherDirOrArchiveRequired", "")
	}
}
