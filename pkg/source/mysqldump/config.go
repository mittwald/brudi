package mysqldump

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "mysqldump"
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

	return config.Validate(c)
}
