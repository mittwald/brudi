package fsbackup_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mittwald/brudi/pkg/restic"
	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/fsbackup"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

const oldPodHost = "brudi-backup-29384710-x7k2p"

type FSBackupSuite struct {
	suite.Suite
}

func (s *FSBackupSuite) SetupTest() {
	commons.TestSetup()
}

func (s *FSBackupSuite) TearDownTest() {
	viper.Reset()
}

func (s *FSBackupSuite) TestResticBackupDirectory() {
	ctx := context.Background()

	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	s.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		if resticErr != nil {
			s.T().Logf("failed to terminate restic container: %v", resticErr)
		}
	}()

	sourceDir := s.T().TempDir()
	restoreDir := s.T().TempDir()
	host := "fsbackup-restic-host"

	filePath := filepath.Join(sourceDir, "sample.txt")
	fileContent := []byte("fsbackup restic content")
	err = os.WriteFile(filePath, fileContent, 0o600)
	s.Require().NoError(err)

	config := createFSBackupConfig(host, sourceDir, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(config))
	s.Require().NoError(err)

	err = source.DoBackupForKind(ctx, fsbackup.Kind, false, true, false, false)
	s.Require().NoError(err)

	err = commons.DoResticRestore(ctx, resticContainer, restoreDir)
	s.Require().NoError(err)

	relativeSourcePath := strings.TrimPrefix(sourceDir, string(os.PathSeparator))
	restoredFilePath := filepath.Join(restoreDir, relativeSourcePath, "sample.txt")

	restoredContent, err := os.ReadFile(restoredFilePath)
	s.Require().NoError(err)
	s.Equal(fileContent, restoredContent)
}

func (s *FSBackupSuite) TestDefaultHostForgetsSnapshotsOfOldPodHosts() {
	snapshotHosts := s.backupAfterOldPodSnapshot("")

	s.Equal([]string{"brudi-fsbackup"}, snapshotHosts)
}

func (s *FSBackupSuite) TestExplicitHostNameKeepsGroupingByHost() {
	snapshotHosts := s.backupAfterOldPodSnapshot("hostName: explicit-host")

	s.ElementsMatch([]string{oldPodHost, "explicit-host"}, snapshotHosts)
}

func (s *FSBackupSuite) backupAfterOldPodSnapshot(hostNameLine string) []string {
	ctx := context.Background()

	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	s.Require().NoError(err)
	defer func() {
		if resticErr := resticContainer.Container.Terminate(ctx); resticErr != nil {
			s.T().Logf("failed to terminate restic container: %v", resticErr)
		}
	}()

	repo := fmt.Sprintf("rest:http://%s:%s/", resticContainer.Address, resticContainer.Port)
	sourceDir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(sourceDir, "sample.txt"), []byte("content"), 0o600))

	_, initErr := runRestic(ctx, "init", "-r", repo)
	s.Require().NoError(initErr)
	_, oldBackupErr := runRestic(ctx, "backup", "-r", repo, "--host", oldPodHost, sourceDir)
	s.Require().NoError(oldBackupErr)

	config := createFSBackupForgetConfig(hostNameLine, sourceDir, repo)
	s.Require().NoError(viper.ReadConfig(bytes.NewBuffer(config)))
	s.Require().NoError(source.DoBackupForKind(ctx, fsbackup.Kind, false, true, true, false))

	snapshotsOut, snapshotsErr := runRestic(ctx, "snapshots", "--json", "-r", repo)
	s.Require().NoError(snapshotsErr)
	var snapshots []restic.Snapshot
	s.Require().NoError(json.Unmarshal(snapshotsOut, &snapshots))

	hosts := make([]string, len(snapshots))
	for index := range snapshots {
		hosts[index] = snapshots[index].Hostname
	}

	return hosts
}

func runRestic(ctx context.Context, args ...string) ([]byte, error) {
	out, runErr := exec.CommandContext(ctx, "restic", args...).Output()
	if runErr != nil {
		return nil, errors.Wrapf(runErr, "restic %s failed: %s", args[0], out)
	}

	return out, nil
}

func TestFSBackupSuite(t *testing.T) {
	suite.Run(t, new(FSBackupSuite))
}

func createFSBackupConfig(host, path, resticIP, resticPort string) []byte {
	return []byte(fmt.Sprintf(
		`
fsbackup:
  options:
    path: %s
  hostName: %s
restic:
  global:
    flags:
      repo: rest:http://%s:%s/
  restore:
    flags:
      path: %s
      target: "/"
    id: "latest"
`, path, host, resticIP, resticPort, path,
	))
}

func createFSBackupForgetConfig(hostNameLine, path, repo string) []byte {
	return []byte(fmt.Sprintf(
		`
fsbackup:
  options:
    path: %s
  %s
restic:
  global:
    flags:
      repo: %s
  forget:
    flags:
      keepLast: 1
`, path, hostNameLine, repo,
	))
}
