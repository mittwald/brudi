package source

import (
	"context"
)

type ExtraResticFlags struct {
	ResticList    bool
	ResticCheck   bool
	ResticPrune   bool
	ResticRebuild bool
	ResticTags    bool
}

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
