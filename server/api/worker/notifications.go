package worker

import (
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/config"
	pushNotifications "github.com/RowenTey/JustJio/server/api/model/push_notifications"
	"github.com/SherClockHolmes/webpush-go"
)

type NotificationData = pushNotifications.NotificationData
type WebPushPayload = pushNotifications.WebPushPayload

func notificationWorker(
	id int,
	notifications <-chan NotificationData,
	vapidEmail, vapidPublicKey, vapidPrivateKey string,
	wg *sync.WaitGroup,
	logger *logrus.Entry,
) {
	defer wg.Done()
	logger.Info("Worker ", id, " started")
	for notification := range notifications {
		webPushPayload := WebPushPayload{
			Title:   notification.Title,
			Message: notification.Message,
		}
		webPushPayloadJson, err := json.Marshal(webPushPayload)
		if err != nil {
			logger.Infof("Worker %d: Error marshalling payload: %s\n", id, err.Error())
			continue
		}

		resp, err := webpush.SendNotification(webPushPayloadJson, notification.Subscription, &webpush.Options{
			// Needed for VAPID authentication to include in token (Safari)
			Subscriber:      vapidEmail,
			VAPIDPublicKey:  vapidPublicKey,
			VAPIDPrivateKey: vapidPrivateKey,
			TTL:             30,
		})
		if err != nil {
			logger.Infof("Worker %d: Error sending notification: %s\n", id, err.Error())
			continue
		}

		logger.Infof("Worker %d: Sent notification! Response: %v\n", id, resp)
		resp.Body.Close()
	}
}

func RunPushNotifications(conf *config.Config, logger *logrus.Logger) chan<- NotificationData {
	pushNotiLogger := logger.WithFields(logrus.Fields{"service": "PushNotificationService"})
	pushNotiLogger.Info("Starting push notification workers...")

	// VAPID keys
	vapidEmail := conf.Vapid.Email
	vapidPublicKey := conf.Vapid.PublicKey
	vapidPrivateKey := conf.Vapid.PrivateKey

	// Buffered channel of 100 notifications
	notifications := make(chan NotificationData, 100)

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		workerLogger := pushNotiLogger.WithFields(logrus.Fields{"worker": i})
		go notificationWorker(i, notifications, vapidEmail, vapidPublicKey, vapidPrivateKey, &wg, workerLogger)
	}

	// Handle SIGINT and SIGTERM signals to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		pushNotiLogger.Info("Received shutdown signal, closing notification channel...")
		close(notifications)

		wg.Wait()
		pushNotiLogger.Info("All workers have finished processing!")
		os.Exit(0)
	}()

	return notifications
}
