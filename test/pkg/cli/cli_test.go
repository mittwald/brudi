package testcli

import (
	"context"
	"github.com/mittwald/brudi/pkg/cli"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

const testFile = "../../testdata/gzip_testfile.txt"
const binary = "gzip"

type CliTestSuite struct {
	suite.Suite
}

func (cliTestSuite *CliTestSuite) TestValidGzipFile() {
	fileName, err := cli.GzipFile(testFile, false)
	defer func() {
		remError := os.Remove(fileName)
		if remError != nil {
			log.WithError(remError).Error("failed to remove test archive from TestValidGzipFile()")
		}
	}()
	cliTestSuite.Require().NoError(err)

	cmd := cli.CommandType{
		Binary: binary,
		Args:   []string{"-t", fileName},
	}
	_, err = cli.Run(context.TODO(), cmd)
	cliTestSuite.Require().NoError(err)
}

func TestCliTestSuite(t *testing.T) {
	suite.Run(t, new(CliTestSuite))
}
