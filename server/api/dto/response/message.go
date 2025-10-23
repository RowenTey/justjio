package response

import "time"

type MessageDto struct {
	ID      uint           `json:"id" binding:"required"`
	RoomID  string         `json:"roomId" binding:"required"`
	Content string         `json:"content" binding:"required"`
	SentAt  time.Time      `json:"sentAt" binding:"required"`
	Sender  MinimalUserDto `json:"sender" binding:"required"`
}

type GetMessagesResponse struct {
	Messages  []MessageDto `json:"messages" binding:"required"`
	Page      int          `json:"page" binding:"required"`
	PageCount int          `json:"pageCount" binding:"required"`
}
