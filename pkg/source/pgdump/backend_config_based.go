package pgdump

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

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
		if strings.TrimSpace(strings.ToLower(config.Options.Flags.Format)) == "directory" {
			return nil, errors.New("the output format of pgdump is set to directory but doPipingBackup " +
				"is also active. Use 'plain', 'custom' or 'tar' instead")
		}
		config.Options.Flags.File = ""
	}

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) (*cli.CommandType, error) {
	gzip := false
	// create temporary, unzipped backup first, thus trim '.gz' extension
	if strings.HasSuffix(b.cfg.Options.Flags.File, cli.GzipSuffix) {
		b.cfg.Options.Flags.File = strings.TrimSuffix(b.cfg.Options.Flags.File, cli.GzipSuffix)
		gzip = true
	}
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

	// zip backup, update flag with the name returned by GzipFile for correct handover to restic
	if gzip {
		b.cfg.Options.Flags.File, err = cli.GzipFile(b.cfg.Options.Flags.File)
		if err != nil {
			return nil, err
		}
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
	return b.cfg.Options.Flags.File
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.Options.Flags.Host
}

func (b *ConfigBasedBackend) CleanUp() error {
	if b.cfg.Options.Flags.Format == "d" {
		return os.RemoveAll(b.GetBackupPath())
	}
	return os.Remove(b.GetBackupPath())
}
