package worker

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	model_push_notifications "github.com/RowenTey/JustJio/model/push_notifications"
	"github.com/RowenTey/JustJio/server-ws/utils"
	"github.com/SherClockHolmes/webpush-go"
)

type NotificationData = model_push_notifications.NotificationData
type WebPushPayload = model_push_notifications.WebPushPayload

func notificationWorker(id int, notifications <-chan NotificationData, vapidEmail, vapidPublicKey, vapidPrivateKey string, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("[PUSH-WORKER] Worker", id, "started")
	for notification := range notifications {
		webPushPayload := WebPushPayload{
			Title:   notification.Title,
			Message: notification.Message,
		}
		webPushPayloadJson, err := json.Marshal(webPushPayload)
		if err != nil {
			log.Printf("[PUSH-WORKER] Worker %d: Error marshalling payload: %s\n", id, err.Error())
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
			log.Printf("[PUSH-WORKER] Worker %d: Error sending notification: %s\n", id, err.Error())
			continue
		}
		log.Printf("[PUSH-WORKER] Worker %d: Sent notification! Response: %v\n", id, resp)
		resp.Body.Close()
	}
}

func RunPushNotification() chan<- NotificationData {
	log.Println("[PUSH-WORKER] Starting push notification workers...")

	// VAPID keys
	vapidEmail := utils.Config("VAPID_EMAIL")
	vapidPublicKey := utils.Config("VAPID_PUBLIC_KEY")
	vapidPrivateKey := utils.Config("VAPID_PRIVATE_KEY")

	notifications := make(chan NotificationData, 100)

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go notificationWorker(i, notifications, vapidEmail, vapidPublicKey, vapidPrivateKey, &wg)
	}

	// Handle SIGINT and SIGTERM signals to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("[PUSH-WORKER] Received shutdown signal, closing notification channel...")
		close(notifications)

		wg.Wait()
		log.Println("[PUSH-WORKER] All workers have finished processing")
		os.Exit(0)
	}()

	return notifications
}
