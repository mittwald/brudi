package fsbackup_test

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
	commons "github.com/mittwald/brudi/test/pkg/source/internal"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

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
