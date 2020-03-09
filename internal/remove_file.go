package internal

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func RemoveAll(filePath string) {
	if err := os.RemoveAll(filePath); err != nil {
		log.WithError(err).WithField("path", filePath).Warn("failed to cleanup backup")
	} else {
		log.WithField("path", filePath).Info("successfully cleaned up backup")
	}
}
