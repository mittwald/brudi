package internal

import (
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

func InitLogger() {
	var logFormatter log.Formatter
	logFormatter = &log.JSONFormatter{}
	if terminal.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("LOG_FORMATTER") != "JSON" {
		logFormatter = &log.TextFormatter{}
	}
	log.SetFormatter(logFormatter)

	log.SetOutput(os.Stdout)

	logLevel := log.InfoLevel
	envLogLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err == nil {
		logLevel = envLogLevel
	}
	log.SetLevel(logLevel)
}
