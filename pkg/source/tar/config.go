package tar

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "tar"
)

type Config struct {
	Options  *Options
	HostName string `validate:"min=1"`
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

	err = viper.UnmarshalKey(fmt.Sprintf("%s.%s.%s", Kind, "options", "paths"), &c.Options.Paths)
	if err != nil {
		return errors.WithStack(err)
	}

	err = viper.UnmarshalKey(fmt.Sprintf("%s.%s", Kind, "hostName"), &c.HostName)
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
