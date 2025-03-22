package util

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func InitLogger(env string) {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	if env == "dev" || env == "staging" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
