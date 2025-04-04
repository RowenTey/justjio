package request

type CreateNotificationRequest struct {
	UserId  uint   `json:"userId"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
