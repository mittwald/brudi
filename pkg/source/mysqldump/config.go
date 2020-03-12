package mysqldump

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "mysqldump"
)

type Config struct {
	Flags *Flags
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(fmt.Sprintf("%s.%s", Kind, "flags"), c.Flags)
	if err != nil {
		return errors.WithStack(err)
	}

	return config.Validate(c)
}
