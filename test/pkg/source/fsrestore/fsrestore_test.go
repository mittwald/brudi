package fsrestore_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/fsbackup"
	"github.com/mittwald/brudi/pkg/source/fsrestore"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type FSRestoreSuite struct {
	suite.Suite
}

func (s *FSRestoreSuite) SetupTest() {
	commons.TestSetup()
}

func (s *FSRestoreSuite) TearDownTest() {
	viper.Reset()
}

func (s *FSRestoreSuite) TestResticRestoreDirectory() {
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
	host := "fsrestore-restic-host"

	filePath := filepath.Join(sourceDir, "sample.txt")
	fileContent := []byte("fsrestore restic content")
	err = os.WriteFile(filePath, fileContent, 0o600)
	s.Require().NoError(err)

	backupConfig := createFSBackupConfig(host, sourceDir, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(backupConfig))
	s.Require().NoError(err)

	err = source.DoBackupForKind(ctx, fsbackup.Kind, false, true, false, false)
	s.Require().NoError(err)

	restoreRoot := s.T().TempDir()

	viper.Reset()
	commons.TestSetup()

	restoreConfig := createFSRestoreConfig(host, sourceDir, restoreRoot, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(restoreConfig))
	s.Require().NoError(err)

	err = source.DoRestoreForKind(ctx, fsrestore.Kind, false, true)
	s.Require().NoError(err)

	relativeSourcePath := strings.TrimPrefix(sourceDir, string(os.PathSeparator))
	restoredFile := filepath.Join(restoreRoot, relativeSourcePath, "sample.txt")

	restoredContent, err := os.ReadFile(restoredFile)
	s.Require().NoError(err)
	s.Equal(fileContent, restoredContent)
}

func TestFSRestoreSuite(t *testing.T) {
	suite.Run(t, new(FSRestoreSuite))
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

func createFSRestoreConfig(host, path, target, resticIP, resticPort string) []byte {
	return []byte(fmt.Sprintf(
		`
fsrestore:
  options:
    path: %s
  hostName: %s
restic:
  global:
    flags:
      repo: rest:http://%s:%s/
  restore:
    flags:
      target: %s
    id: "latest"
`, path, host, resticIP, resticPort, target,
	))
}
