package logger

import (
	"os"

	"github.com/RowenTey/JustJio/server/api/pkg/app"
	log "github.com/sirupsen/logrus"
)

func InitLogger(ctx *app.Context) *log.Logger {
	env := ctx.Config.Environment

	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{})
	logger.SetOutput(os.Stdout)

	logger.SetLevel(log.InfoLevel)
	if env != "production" {
		logger.SetLevel(log.DebugLevel)
	}

	return logger
}
