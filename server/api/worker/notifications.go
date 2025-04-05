package worker

import (
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/config"
	model_push_notifications "github.com/RowenTey/JustJio/server/api/model/push_notifications"
	"github.com/SherClockHolmes/webpush-go"
)

type NotificationData = model_push_notifications.NotificationData
type WebPushPayload = model_push_notifications.WebPushPayload

func notificationWorker(
	id int,
	notifications <-chan NotificationData,
	vapidEmail, vapidPublicKey, vapidPrivateKey string,
	wg *sync.WaitGroup,
	logger *log.Entry,
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

func RunPushNotification() chan<- NotificationData {
	logger := log.WithFields(log.Fields{"service": "PushNotificationService"})

	logger.Info("Starting push notification workers...")

	// VAPID keys
	vapidEmail := config.Config("VAPID_EMAIL")
	vapidPublicKey := config.Config("VAPID_PUBLIC_KEY")
	vapidPrivateKey := config.Config("VAPID_PRIVATE_KEY")

	notifications := make(chan NotificationData, 100)

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		workerLogger := logger.WithFields(log.Fields{"service": "PushNotificationService", "worker": i})
		go notificationWorker(i, notifications, vapidEmail, vapidPublicKey, vapidPrivateKey, &wg, workerLogger)
	}

	// Handle SIGINT and SIGTERM signals to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, closing notification channel...")
		close(notifications)

		wg.Wait()
		logger.Info("All workers have finished processing!")
		os.Exit(0)
	}()

	return notifications
}
