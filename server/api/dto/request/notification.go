package request

type CreateNotificationRequest struct {
	UserId  uint   `json:"userId" validate:"required"`
	Title   string `json:"title" validate:"required,min=1,max=100"`
	Content string `json:"content" validate:"required,min=1,max=500"`
}
