package mysqldump

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mittwald/brudi/pkg/cli"
)

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

	return &ConfigBasedBackend{cfg: config}, nil
}

func (b *ConfigBasedBackend) CreateBackup(ctx context.Context) error {
	gzip := false
	// create temporary, unzipped backup first, thus trim '.gz' extension
	if strings.HasSuffix(b.cfg.Options.Flags.ResultFile, cli.GzipSuffix) {
		b.cfg.Options.Flags.ResultFile = strings.TrimSuffix(b.cfg.Options.Flags.ResultFile, cli.GzipSuffix)
		gzip = true
	}

	cmd := cli.CommandType{
		Binary: binary,
		Args:   cli.StructToCLI(b.cfg.Options),
	}
	log.Warnf("cmd: %+v", cmd)
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return errors.WithStack(fmt.Errorf("%+v - %s", err, out))
	}

	// zip backup, update flag with the name returned by GzipFile for correct handover to restic
	if gzip {
		b.cfg.Options.Flags.ResultFile, err = cli.GzipFile(b.cfg.Options.Flags.ResultFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *ConfigBasedBackend) GetBackupPath() string {
	return b.cfg.Options.Flags.ResultFile
}

func (b *ConfigBasedBackend) GetHostname() string {
	return b.cfg.Options.Flags.Host
}

func (b *ConfigBasedBackend) CleanUp() error {
	return os.Remove(b.GetBackupPath())
}
