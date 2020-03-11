package tar

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
)

const (
	Kind = "tar"
)

type Config struct {
	Options  *Options
	HostName string
}

func (c *Config) InitFromViper() error {
	err := viper.UnmarshalKey(Kind, &c.Options)
	if err != nil {
		return errors.WithStack(err)
	}

	c.HostName, err = os.Hostname()
	if err != nil {
		return errors.WithStack(err)
	}

	if c.Options.Flags.File == "" {
		return errors.WithStack(fmt.Errorf("no tar output path provided"))
	}

	return nil
}
