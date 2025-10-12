package request

type CreateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=2000"`
}
