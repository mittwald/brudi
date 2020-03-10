package mongodump

import (
	"github.com/mongodb/mongo-tools-common/options"
	"github.com/mongodb/mongo-tools/mongodump"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "mongodump"
)

type Config struct {
	ToolOptions   *options.ToolOptions
	InputOptions  *mongodump.InputOptions
	OutputOptions *mongodump.OutputOptions
}

func (c *Config) InitFromViper() error {
	err := viper.UnmarshalKey(Kind, c.ToolOptions)
	if err != nil {
		return errors.WithStack(err)
	}

	err = viper.UnmarshalKey(
		config.GenerateConfigKey(Kind, "inputOptions"),
		c.InputOptions,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	err = viper.UnmarshalKey(
		config.GenerateConfigKey(Kind, "outputOptions"),
		c.OutputOptions,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	err = viper.UnmarshalKey(
		config.GenerateConfigKey(Kind, "auth.username"),
		&c.ToolOptions.Auth.Username,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	err = viper.UnmarshalKey(
		config.GenerateConfigKey(Kind, "auth.password"),
		&c.ToolOptions.Auth.Password,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	if c.OutputOptions.NumParallelCollections == 0 {
		c.OutputOptions.NumParallelCollections = 1
	}

	return nil
}
