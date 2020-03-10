package mysqldump

import (
	"crypto/sha256"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

var testConfSHA256 = "8c3cf10504b5b401d9954bb266519bfdd9c53f3a50023471ccce6e8c47902987"

func TestGenerateMyCnfFromClientOptions(t *testing.T) {
	cfg := Config{
		ClientOptions: map[string]interface{}{
			"opt":  true,
			"host": "localhorst",
			"port": 1234,
		},
		clientMyCnfPath: "",
	}
	assert.NoError(t, cfg.generateClientMyCnf())

	f, err := os.Open(cfg.clientMyCnfPath)
	assert.NoError(t, err)

	defer f.Close()
	defer os.RemoveAll(cfg.clientMyCnfPath)

	h := sha256.New()
	_, err = io.Copy(h, f)
	assert.NoError(t, err)

	assert.Equal(t, testConfSHA256, fmt.Sprintf("%x", h.Sum(nil)))
}
