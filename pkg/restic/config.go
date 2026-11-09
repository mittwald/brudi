package restic

import (
	"github.com/mittwald/brudi/pkg/config"
)

const (
	Kind = "restic"

	defaultHostPrefix   = "brudi-"
	defaultForgetGroups = "paths"
)

type Config struct {
	Global  *GlobalOptions
	Backup  *BackupOptions
	Forget  *ForgetOptions
	Restore *RestoreOptions
}

func (c *Config) InitFromViper() error {
	err := config.InitializeStructFromViper(Kind, c)
	if err != nil {
		return err
	}

	return config.Validate(c)
}

// applyHostDefaults keeps an explicitly configured host untouched. Without one, the
// snapshot host would be the pod name, which changes on every Kubernetes CronJob run,
// so every snapshot would land in its own forget group and retention would never delete
// anything. A stable host plus grouping by paths keeps retention working, including for
// snapshots written with the old pod-name hosts.
func (c *Config) applyHostDefaults(kind, sourceHostname string) {
	if c.Backup.Flags.Host != "" {
		return
	}

	if sourceHostname != "" {
		c.Backup.Flags.Host = sourceHostname

		return
	}

	c.Backup.Flags.Host = defaultHostPrefix + kind
	if c.Forget.Flags.GroupBy == "" {
		c.Forget.Flags.GroupBy = defaultForgetGroups
	}
}
