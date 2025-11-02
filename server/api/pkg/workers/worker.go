package workers

import (
	"sync"

	"github.com/RowenTey/JustJio/server/api/pkg/app"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/notifications"
	"github.com/go-co-op/gocron"
)

type WorkersDependencies struct {
	NotificationsChan chan<- notifications.NotificationData
	WorkersWg         *sync.WaitGroup
	Scheduler         *gocron.Scheduler
}

func StartWorkers(ctx *app.Context) *WorkersDependencies {
	logger := ctx.Logger
	logger.Info("Starting workers module...")

	scheduler := startMaterializedViewRefresher(ctx, "user_non_friends")
	notificationsChan, notificationsWg := RunPushNotifications(ctx)

	// // Handle SIGINT and SIGTERM signals to gracefully shutdown
	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// go func() {
	// 	<-sigChan
	// 	logger.Info("Received shutdown signal, closing workers...")

	// 	close(notificationsChan)
	// 	notificationsWg.Wait()
	// 	scheduler.Stop()

	// 	logger.Info("All workers have finished processing!")
	// 	os.Exit(0)
	// }()

	return &WorkersDependencies{
		NotificationsChan: notificationsChan,
		WorkersWg:         notificationsWg,
		Scheduler:         scheduler,
	}
}
