package restic

import (
	"fmt"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "restic"
)

type Config struct {
	Repository      string `env:"RESTIC_REPOSITORY"`
	Password        string `env:"RESTIC_PASSWORD" validate:"required,min=1"`
	BucketName      string
	Host            string `validate:"required,min=1"`
	AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	Region          string
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(Kind, c)
	if err != nil {
		return err
	}

	if c.Repository == "" {
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

	return config.Validate(c)
}

func (c *Config) EnsureEnv() error {
	v := reflect.ValueOf(*c)

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("env")

		if tag == "" {
			continue
		}

		if os.Getenv(tag) == "" {
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
