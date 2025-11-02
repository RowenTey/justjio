package response

import "time"

type NotificationDto struct {
	ID        uint      `json:"id" binding:"required"`
	Title     string    `json:"title" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	IsRead    bool      `json:"isRead" binding:"required"`
	CreatedAt time.Time `json:"createdAt" binding:"required"`
}
