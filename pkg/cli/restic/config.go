package restic

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/mittwald/brudi/pkg/config"
	log "github.com/sirupsen/logrus"
	"os"
	"reflect"
)

const (
	Kind = "restic"
)

type Config struct {
	Repository      string `env:"RESTIC_REPOSITORY"`
	Password        string `env:"RESTIC_PASSWORD" validate:"required,min=1"`
	BucketName      string `validate:"required,min=1"`
	Host            string `validate:"required,min=1"`
	AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	Region          string `validate:"required,min=1"`
}

// viper is sadly not capable of resolving childs env when fetching the parent key
// therefore we have to workaround the env-resolving
// https://github.com/spf13/viper/issues/696
func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(Kind, c)
	if err != nil {
		return err
	}

	if len(c.Repository) == 0 {
		c.Repository = fmt.Sprintf(
			"s3:%s/%s/%s",
			c.Region,
			c.BucketName,
			c.Host,
		)
	}

	err = c.EnsureEnv()
	if err != nil {
		return err
	}

	validate := validator.New()
	return validate.Struct(c)
}

func (c *Config) EnsureEnv() error {
	v := reflect.ValueOf(*c)

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("env")

		if tag == "" {
			continue
		}

		if len(os.Getenv(tag)) == 0 {
			err := os.Setenv(tag, v.Field(i).String())
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"key":   tag,
				"value": v.Field(i).String(),
			}).Debug("environment variable for restic set")
		}
	}

	return nil
}
