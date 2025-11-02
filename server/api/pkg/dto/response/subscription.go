package response

type SubscriptionDto struct {
	ID       string `json:"id" binding:"required"`
	UserID   uint   `json:"userId" binding:"required"`
	Endpoint string `json:"endpoint" binding:"required"`
	Auth     string `json:"auth" binding:"required"`
	P256dh   string `json:"p256dh" binding:"required"`
}
