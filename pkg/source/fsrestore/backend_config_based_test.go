package fsrestore

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
fsrestore:
  options:
    path: %s
  hostName: %s
`, path, host,
	)

	err := viper.ReadConfig(bytes.NewBufferString(config))
	require.NoError(t, err)
}

func TestRestoreBackup_NoOp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	host, err := os.Hostname()
	require.NoError(t, err)

	loadConfig(t, dir, host)

	backend, err := NewConfigBasedBackend()
	require.NoError(t, err)

	err = backend.RestoreBackup(context.Background())
	require.NoError(t, err)
	require.Equal(t, dir, backend.GetBackupPath())
	require.Equal(t, host, backend.GetHostname())
}

func TestCleanUp_NoOp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	loadConfig(t, dir, "fsrestore-cleanup-host")

	backend, err := NewConfigBasedBackend()
	require.NoError(t, err)

	err = backend.CleanUp()
	require.NoError(t, err)

	_, statErr := os.Stat(dir)
	require.NoError(t, statErr)
}
