package response

import (
	"github.com/RowenTey/JustJio/model"
)

type JoinRoomResponse struct {
	Room     model.Room   `json:"room"`
	Attendes []model.User `json:"attendees"`
}

type CreateRoomResponse struct {
	Room    model.Room         `json:"room"`
	Invites []model.RoomInvite `json:"invites"`
}
