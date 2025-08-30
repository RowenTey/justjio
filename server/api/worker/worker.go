package worker

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func StartWorkers(conf *config.Config, logger *logrus.Logger, dbConn *gorm.DB) chan<- NotificationData {
	scheduler := startMaterializedViewRefresher(dbConn, logger, "user_non_friends")
	notificationsChan, notificationsWg := runPushNotifications(conf, logger)

	// Handle SIGINT and SIGTERM signals to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, closing workers...")

		close(notificationsChan)
		notificationsWg.Wait()
		scheduler.Stop()

		logger.Info("All workers have finished processing!")
		os.Exit(0)
	}()

	return notificationsChan
}
