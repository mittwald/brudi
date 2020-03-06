package backend

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type ConfigHost struct {
	Hostname string `validate:"required"`
	Port     int    `validate:"gte=1,lte=65535"`
}

type ConfigAuth struct {
	Username string
	Password string
}

// viper is sadly not capable of resolving childs env when fetching the parent key
// therefore we have to workaround the env-resolving
// https://github.com/spf13/viper/issues/696
func (cA *ConfigAuth) InitFromViper() error {
	err := cA.InitUsernameFromViper()
	if err != nil {
		return err
	}

	return cA.InitPasswordFromViper()
}

func (cA *ConfigAuth) InitUsernameFromViper() error {
	return errors.WithStack(
		viper.UnmarshalKey("source.auth.username", &cA.Username),
	)
}

func (cA *ConfigAuth) InitPasswordFromViper() error {
	return errors.WithStack(
		viper.UnmarshalKey("source.auth.password", &cA.Password),
	)
}
