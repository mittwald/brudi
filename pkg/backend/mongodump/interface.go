package mongodump

type Backend interface {
	CreateBackup() error
	GetBackupPath() string
	GetHostname() string
}

// maybe there will be another way to configure mongodump backend in the future
// for example via arguments instead of yaml-config
func NewBackend() (Backend, error) {
	return newConfigBasedBackend()
}
