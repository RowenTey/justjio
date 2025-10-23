package response

import "time"

type MinimalUserDto struct {
	ID         uint   `json:"id" binding:"required"`
	Username   string `json:"username" binding:"required"`
	PictureUrl string `json:"pictureUrl" binding:"required"`
}

type FriendRequestDto struct {
	ID          uint           `json:"id" binding:"required"`
	Status      string         `json:"status" binding:"required"`
	SentAt      time.Time      `json:"sentAt" binding:"required"`
	RespondedAt time.Time      `json:"respondedAt,omitempty"`
	Sender      MinimalUserDto `json:"sender" binding:"required"`
	Receiver    MinimalUserDto `json:"receiver" binding:"required"`
}
