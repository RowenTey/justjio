package workers

import (
	"encoding/json"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/pkg/app"
	"github.com/RowenTey/JustJio/server/api/pkg/config"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/notifications"
	"github.com/SherClockHolmes/webpush-go"
)

func RunPushNotifications(ctx *app.Context) (chan<- notifications.NotificationData, *sync.WaitGroup) {
	log := ctx.Logger.WithFields(logrus.Fields{"component": "PushNotificationWorker"})
	log.Info("Starting push notification workers...")

	// Buffered channel of 100 notifications
	notifications := make(chan notifications.NotificationData, 100)

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go notificationWorker(i, notifications, &ctx.Config.Vapid, &wg, log)
	}

	return notifications, &wg
}

func notificationWorker(
	id int,
	notificationsChan <-chan notifications.NotificationData,
	vapidConfig *config.VapidConfig,
	wg *sync.WaitGroup,
	logger *logrus.Entry,
) {
	defer wg.Done()

	log := logger.WithFields(logrus.Fields{"worker": id})

	// VAPID keys
	vapidEmail := vapidConfig.Email
	vapidPublicKey := vapidConfig.PublicKey
	vapidPrivateKey := vapidConfig.PrivateKey

	log.Info("Worker started")
	for notification := range notificationsChan {
		webPushPayload := notifications.WebPushPayload{
			Title:   notification.Title,
			Message: notification.Message,
		}
		webPushPayloadJson, err := json.Marshal(webPushPayload)
		if err != nil {
			log.Errorf("Error marshalling payload: %s\n", err.Error())
			continue
		}

		resp, err := webpush.SendNotification(
			webPushPayloadJson,
			notification.Subscription,
			&webpush.Options{
				// Needed for VAPID authentication to include in token (Safari)
				Subscriber:      vapidEmail,
				VAPIDPublicKey:  vapidPublicKey,
				VAPIDPrivateKey: vapidPrivateKey,
				TTL:             30,
			})
		if err != nil {
			log.Errorf("Error sending notification: %s\n", err.Error())
			continue
		}

		log.Infof("Sent notification! Response: %v\n", resp)
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Failed to close response body: %v\n", err)
		}
	}
}
