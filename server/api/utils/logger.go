package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func InitLogger(env string) *log.Logger {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{})
	logger.SetOutput(os.Stdout)

	if env == "dev" || env == "staging" {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}
	return logger
}

func AddServiceField(logger *log.Logger, serviceName string) *log.Entry {
	return logger.WithField("service", serviceName)
}
