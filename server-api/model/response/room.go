package response

import (
	"github.com/RowenTey/JustJio/model"
)

type GetNumRoomsResponse struct {
	Count int `json:"count"`
}

type JoinRoomResponse struct {
	Room      model.Room   `json:"room"`
	Attendees []model.User `json:"attendees"`
}

type CreateRoomResponse struct {
	Room    model.Room         `json:"room"`
	Invites []model.RoomInvite `json:"invites"`
}
