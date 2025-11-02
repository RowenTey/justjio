package request

type CreateSubscriptionRequest struct {
	UserID   uint   `json:"userId" validate:"required" example:"1"`
	Endpoint string `json:"endpoint" validate:"required,url" example:"https://fcm.googleapis.com/fcm/send/abc123"`
	Auth     string `json:"auth" validate:"required" example:"BCVxsr7N_eNgVRqvHtD0zTZsEc6-VV-JvLexhqUzORcx"`
	P256dh   string `json:"p256dh" validate:"required" example:"BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtUbVlUls0VJXg7A8u"`
}
