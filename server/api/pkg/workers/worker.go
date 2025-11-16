package workers

import (
	"sync"

	"github.com/RowenTey/JustJio/server/api/pkg/app"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/notifications"
	"github.com/go-co-op/gocron/v2"
)

type WorkersModule struct {
	NotificationsChan chan<- notifications.NotificationData
	WorkersWg         *sync.WaitGroup
	Scheduler         *gocron.Scheduler
}

func StartWorkers(ctx *app.Context) *WorkersModule {
	logger := ctx.Logger
	logger.Info("Starting workers module...")

	scheduler := startMaterializedViewRefresher(ctx, "user_non_friends")
	notificationsChan, notificationsWg := RunPushNotifications(ctx)

	return &WorkersModule{
		NotificationsChan: notificationsChan,
		WorkersWg:         notificationsWg,
		Scheduler:         scheduler,
	}
}
