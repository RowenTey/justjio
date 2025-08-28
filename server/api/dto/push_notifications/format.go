package model_push_notifications

import "github.com/SherClockHolmes/webpush-go"

type NotificationData struct {
	Subscription *webpush.Subscription
	Title        string
	Message      string
}

type WebPushPayload struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}
