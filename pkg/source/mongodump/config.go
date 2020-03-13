package mongodump

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/go-playground/validator/v10"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "mongodump"
)

type Config struct {
	Options *Options
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(fmt.Sprintf("%s.%s", Kind, config.KeyOptionsFlags), c.Options.Flags)
	if err != nil {
		return errors.WithStack(err)
	}

	err = viper.UnmarshalKey(fmt.Sprintf("%s.%s", Kind, config.KeyOptionsAdditionalArgs), &c.Options.AdditionalArgs)
	if err != nil {
		return errors.WithStack(err)
	}

	return config.Validate(c, configStructLevelValidation)
}

func configStructLevelValidation(sl validator.StructLevel) {
	c := sl.Current().Interface().(Config)

	if c.Options.Flags.Out == "" && c.Options.Flags.Archive == "" {
		sl.ReportError(c.Options.Flags.Out, "out", "Out", "eitherOutOrArchiveRequired", "")
		sl.ReportError(c.Options.Flags.Archive, "archive", "Archive", "eitherOutOrArchiveRequired", "")
	}
}
