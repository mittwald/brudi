package restic

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HostDefaultsSuite struct {
	suite.Suite
}

func newConfigWith(backupHost, groupBy string) *Config {
	return &Config{
		Backup: &BackupOptions{Flags: &BackupFlags{Host: backupHost}},
		Forget: &ForgetOptions{Flags: &ForgetFlags{GroupBy: groupBy}},
	}
}

func (s *HostDefaultsSuite) TestWithoutAnyHostUsesStableHostAndGroupsByPaths() {
	conf := newConfigWith("", "")

	conf.applyHostDefaults("fsbackup", "")

	s.Equal("brudi-fsbackup", conf.Backup.Flags.Host)
	s.Equal("paths", conf.Forget.Flags.GroupBy)
}

func (s *HostDefaultsSuite) TestWithoutAnyHostKeepsExplicitGroupBy() {
	conf := newConfigWith("", "host,paths")

	conf.applyHostDefaults("tar", "")

	s.Equal("brudi-tar", conf.Backup.Flags.Host)
	s.Equal("host,paths", conf.Forget.Flags.GroupBy)
}

func (s *HostDefaultsSuite) TestSourceHostnameIsUsedWithResticDefaultGrouping() {
	conf := newConfigWith("", "")

	conf.applyHostDefaults("fsbackup", "my-host")

	s.Equal("my-host", conf.Backup.Flags.Host)
	s.Empty(conf.Forget.Flags.GroupBy)
}

func (s *HostDefaultsSuite) TestResticBackupHostWinsOverSourceHostname() {
	conf := newConfigWith("restic-host", "")

	conf.applyHostDefaults("mysqldump", "db.example.com")

	s.Equal("restic-host", conf.Backup.Flags.Host)
	s.Empty(conf.Forget.Flags.GroupBy)
}

func TestHostDefaultsSuite(t *testing.T) {
	suite.Run(t, new(HostDefaultsSuite))
}
