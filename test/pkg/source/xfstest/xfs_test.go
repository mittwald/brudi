package xfs_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/mittwald/brudi/pkg/source"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"
)

const dumpName = "../../../../testbackup.xfsdump"
const loopDeviceName = "/dev/loop10"
const mountPoint = "../../../../xfsmount"

type XFSTestSuite struct {
	suite.Suite
}

func (xfsTestSuite *XFSTestSuite) SetupTest() {
	commons.TestSetup()
}

// TearDownTest resets viper after a test
func (xfsTestSuite *XFSTestSuite) TearDownTest() {
	viper.Reset()
}

func (xfsTestSuite *XFSTestSuite) TestBasicXFSDump() {
	ctx := context.Background()

	tarConfig := createXFSConfig("", "")
	err := viper.ReadConfig(bytes.NewBuffer(tarConfig))
	xfsTestSuite.Require().NoError(err)

	dirName := fmt.Sprintf("%s/testdir", mountPoint)

	os.Mkdir(dirName, 744)
	xfsTestSuite.Require().NoError(err)

	err = source.DoBackupForKind(ctx, "xfsdump", false, false, false)
	xfsTestSuite.Require().NoError(err)

	err = os.Remove(dirName)
	xfsTestSuite.Require().NoError(err)

	err = source.DoRestoreForKind(ctx, "xfsrestore", false, false, false)
	xfsTestSuite.Require().NoError(err)

	_, err = os.Stat(dirName)
	xfsTestSuite.Require().NoError(err)

	err = os.Remove(dumpName)
	xfsTestSuite.Require().NoError(err)

	err = os.Remove(dirName)
	xfsTestSuite.Require().NoError(err)
}

func (xfsTestSuite *XFSTestSuite) TestXFSDumpRestic() {
	ctx := context.Background()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	xfsTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		if resticErr != nil {
			log.WithError(resticErr).Error("failed to terminate xfs restic container")
		}
	}()

	xfsConfig := createXFSConfig(resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(xfsConfig))
	xfsTestSuite.Require().NoError(err)

	dirName := fmt.Sprintf("%s/testdir", mountPoint)

	err = os.Mkdir(dirName, 744)
	xfsTestSuite.Require().NoError(err)

	err = source.DoBackupForKind(ctx, "xfsdump", false, true, false)
	xfsTestSuite.Require().NoError(err)

	err = os.Remove(dirName)
	xfsTestSuite.Require().NoError(err)

	err = source.DoRestoreForKind(ctx, "xfsrestore", false, true, false)
	xfsTestSuite.Require().NoError(err)

	_, err = os.Stat(dirName)
	xfsTestSuite.Require().NoError(err)
	err = os.Remove(dumpName)
	xfsTestSuite.Require().NoError(err)
	err = os.Remove(dirName)
	xfsTestSuite.Require().NoError(err)
}

func TestXFSTestSuite(t *testing.T) {
	suite.Run(t, new(XFSTestSuite))
}

// createXFSConfig creates a brudi config for the xfs commands
func createXFSConfig(resticIP, resticPort string) []byte {
	return []byte(fmt.Sprintf(`
xfsdump:
  options:
    flags:
      level: 0
      destination: %s
    additionalArgs: []
    targetFS: %s
xfsrestore:
  options:
    flags:
      source: %s
    additionalArgs: []
    destFS: %s
restic:
  global:
    flags:
      repo: rest:http://%s:%s/
  restore:
    flags:
      target: "/"
    id: "latest"
`, dumpName, loopDeviceName, dumpName, mountPoint, resticIP, resticPort))
}
