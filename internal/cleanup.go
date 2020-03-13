package internal

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ClearDirectory(pathToClear string, fileSuffix ...string) error {
	// delete desired files from directory
	err := filepath.Walk(pathToClear, getDeleteSuffixedFileWalkFunc(fileSuffix...))
	if err != nil {
		return err
	}

	// delete parent directory
	return os.Remove(pathToClear)
}

func getDeleteSuffixedFileWalkFunc(fileSuffix ...string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		parentDir := filepath.Dir(path)
		fieldLogger := log.WithFields(
			log.Fields{
				"parentDir": parentDir,
				"path":      path,
				"task":      "cleanupFiles",
			},
		)

		if err != nil {
			fieldLogger.
				WithError(err).
				Debug("encountered error while walking path")
			return err
		}

		if info.IsDir() {
			fieldLogger.
				Debug("skipping directory")
			return nil
		}

		fieldLogger.Debug("processing file")

		validFile := false

		if len(fileSuffix) == 0 {
			validFile = true
		} else {
			for i := range fileSuffix {
				if strings.HasSuffix(path, fileSuffix[i]) {
					validFile = true
				}
			}
		}

		if !validFile {
			fieldLogger.Debug("skipping file due to invalid suffix")
			return nil
		}

		removeErr := os.Remove(path)
		if removeErr != nil {
			return err
		}

		removeErr = os.Remove(parentDir)
		if removeErr != nil && strings.Contains(removeErr.Error(), "directory not empty") {
			fieldLogger.Debug("directory is not empty")
			return nil
		} else if removeErr != nil {
			return err
		}

		fieldLogger.Debug("parent directory deleted")

		return nil
	}
}
