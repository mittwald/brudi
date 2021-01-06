package internal

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

const ResticPort = "8000/tcp"

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
		"RESTIC_PASSWORD": "mongorepo",
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
	os.Setenv("RESTIC_PASSWORD", "mongorepo")
}

func DoResticRestore(ctx context.Context, resticContainer TestContainerSetup, dataDir string) error {
	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", fmt.Sprintf("rest:http://%s:%s/",
		resticContainer.Address, resticContainer.Port),
		"--target", dataDir, "latest")
	_, err := cmd.CombinedOutput()
	return err
}
