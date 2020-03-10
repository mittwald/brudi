package mysqldump

import "context"

type Backend interface {
	CreateBackup(ctx context.Context) error
	GetBackupPath() string
	GetHostname() string
}

// maybe there will be another way to configure mysqldump backend in the future
// for example via arguments instead of yaml-config
func NewBackend() (Backend, error) {
	return newConfigBasedBackend()
}
