package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

const ResticPort = "8000/tcp"
const ResticPassword = "resticRepo"

// TestContainerSetup is a wrapper for testcontainers that gives easy access to container-address and container-port
type TestContainerSetup struct {
	Container testcontainers.Container
	Address   string
	Port      string
}

// ResticReq is a testcontainers request for a restic container
var ResticReq = testcontainers.ContainerRequest{
	Image:        "restic/rest-server:latest",
	ExposedPorts: []string{ResticPort},
	Env: map[string]string{
		"OPTIONS":         "--no-auth",
		"RESTIC_PASSWORD": ResticPassword,
	},
}

// NewTestContainerSetup creates a TestContainerSetup which acts as a wrapper for the testcontainer specified by request
func NewTestContainerSetup(ctx context.Context, request *testcontainers.ContainerRequest, port nat.Port) (TestContainerSetup, error) {
	result := TestContainerSetup{}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: *request,
		Started:          true,
	})
	if err != nil {
		return TestContainerSetup{}, err
	}
	result.Container = container
	contPort, err := container.MappedPort(ctx, port)
	if err != nil {
		return TestContainerSetup{}, err
	}
	result.Port = fmt.Sprint(contPort.Int())
	host, err := container.Host(ctx)
	if err != nil {
		return TestContainerSetup{}, err
	}
	result.Address = host

	return result, nil
}

// TestSetup resets Viper and then performs initialization
func TestSetup() {
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	os.Setenv("RESTIC_PASSWORD", ResticPassword)
}

// GetProgramVersion tries to run the given program with the given version argument to determine its version.
// Leave the version string empty to use "--version".
func GetProgramVersion(program, versionArg string) (string, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	if versionArg == "" {
		versionArg = "--version"
	}
	cmd := exec.CommandContext(ctx, program, versionArg)
	version, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "error running '%s %s'", program, versionArg)
	}
	return string(version), nil
}

// GetProgramsVersions does the same as GetProgramVersion but for multiple programs. Give the programs and their versions
// like this: "[program]", "[version]", "[program]", "[version]"... - Leave the version string empty to use "--version".
// Returns after all programs have been tested.
func GetProgramsVersions(programsAndVersions ...string) (versions []string, err error) {
	versions = make([]string, 0, len(programsAndVersions))
	errs := make([]string, 0, len(programsAndVersions))
	for i := 0; i < len(programsAndVersions); i += 2 {
		v, err := GetProgramVersion(programsAndVersions[i], programsAndVersions[i+1])
		versions = append(versions, v)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	err = nil
	if len(errs) > 0 {
		err = errors.Errorf("got error(s) while determining the versions of required programsAndVersions for testing:\n\t%s",
			strings.Join(errs, "\n\t"))
	}
	return
}

// DoResticRestore pulls the given backup from the given restic repo
func DoResticRestore(ctx context.Context, resticContainer TestContainerSetup, dataDir string) error {
	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", // nolint: gosec
		fmt.Sprintf("rest:http://%s:%s/",
			resticContainer.Address, resticContainer.Port),
		"--target", dataDir, "latest")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Errorf("failed to execute restic restore: \n Output: %s \n Error: %s", out, err)
	}
	return err
}
