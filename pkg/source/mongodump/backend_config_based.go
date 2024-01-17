package mongodump

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/mittwald/brudi/internal"
	"github.com/mittwald/brudi/pkg/cli"
)

//var _ source.Generic = &ConfigBasedBackend{}

type ConfigBasedBackend struct {
	cfg *Config
}

func NewConfigBasedBackend() (*ConfigBasedBackend, error) {
	config := &Config{
		&Options{
			Flags:          &Flags{},
			AdditionalArgs: []string{},
		},
	}

	err := config.InitFromViper()
	if err != nil {
		return nil, err
	}
	if viper.GetBool(cli.DoStdinBackupKey) {
		config.Options.Flags.Archive = ""
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) (*cli.CommandType, error) {
	cmd := b.GetBackupCommand()

	var out []byte
	var err error = nil
	if viper.GetBool(cli.DoStdinBackupKey) {
		cmd.PipeReady = &sync.Cond{L: &sync.Mutex{}}
		go func() {
			_, err = cli.Run(ctx, &cmd, true)
			if err != nil {
				log.Errorf("error while running backup program: %v", err)
			}
		}()
		cmd.PipeReady.L.Lock()
		cmd.PipeReady.Wait()
		cmd.PipeReady.L.Unlock()
		return &cmd, err
	} else {
		out, err = cli.Run(ctx, &cmd, false)
	}
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("%+v - %s", err, out))
	}

	return nil, nil
}

func (b *ConfigBasedBackend) GetBackupCommand() cli.CommandType {
	return cli.CommandType{
		Binary: binary,
		Args:   cli.StructToCLI(b.cfg.Options),
	}
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	if b.cfg.Options.Flags.Archive != "" {
		return b.cfg.Options.Flags.Archive
	}

	return b.cfg.Options.Flags.Out
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.Options.Flags.Host
}

func (b *ConfigBasedBackend) CleanUp() error {
	var fileTypes []string

	if b.cfg.Options.Flags.Archive != "" {
		return os.Remove(b.cfg.Options.Flags.Archive)
	} else if b.cfg.Options.Flags.Gzip {
		fileTypes = append(fileTypes, ".bson.gz", ".json.gz")
	} else {
		fileTypes = append(fileTypes, ".bson", ".json")
	}

	return internal.ClearDirectory(b.GetBackupPath(), fileTypes...)
}
