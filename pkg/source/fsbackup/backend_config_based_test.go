package fsbackup

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func loadConfig(t *testing.T, path, host string) {
	t.Helper()

	viper.Reset()
	viper.SetConfigType("yaml")
	config := fmt.Sprintf(
		`
fsbackup:
  options:
    path: %s
  hostName: %s
`, path, host,
	)

	err := viper.ReadConfig(bytes.NewBufferString(config))
	require.NoError(t, err)
}

func TestCreateBackup_PathExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	host := "fsbackup-test-host"

	loadConfig(t, dir, host)

	backend, err := NewConfigBasedBackend()
	require.NoError(t, err)

	err = backend.CreateBackup(context.Background())
	require.NoError(t, err)

	require.Equal(t, dir, backend.GetBackupPath())
	require.Equal(t, host, backend.GetHostname())
}

func TestCreateBackup_PathMissing(t *testing.T) {
	t.Parallel()

	missingDir := filepath.Join(t.TempDir(), "missing")
	host := "fsbackup-missing-test"

	loadConfig(t, missingDir, host)

	backend, err := NewConfigBasedBackend()
	require.NoError(t, err)

	err = backend.CreateBackup(context.Background())
	require.Error(t, err)
}

func TestCleanUp_NoOp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	loadConfig(t, dir, "fsbackup-cleanup-host")

	backend, err := NewConfigBasedBackend()
	require.NoError(t, err)

	err = backend.CreateBackup(context.Background())
	require.NoError(t, err)

	err = backend.CleanUp()
	require.NoError(t, err)

	_, statErr := os.Stat(dir)
	require.NoError(t, statErr)
}
