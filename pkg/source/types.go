package source

import (
	"context"
)

type Generic interface {
	CreateBackup(ctx context.Context) error
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
