package source

import (
	"context"
	"github.com/mittwald/brudi/pkg/cli"
)

type Generic interface {
	CreateBackup(ctx context.Context) (*cli.CommandType, error) // TODO: Remove *cli.CommandType when --stdin-command was added to restic
	GetBackupCommand() cli.CommandType
	GetBackupPath() string
	GetHostname() string
	CleanUp() error
}

type GenericRestore interface {
	RestoreBackup(ctx context.Context) error
	GetBackupPath() string
	GetHostname() string
	CleanUp() error
}
