package request

import (
	"github.com/RowenTey/JustJio/model"

	"gorm.io/datatypes"
)

type CreateRoomRequest struct {
	Room       model.Room     `json:"room"`
	InviteesId datatypes.JSON `json:"invitees" swaggertype:"array,string"`
	Message    string         `json:"message"`
}

type RespondToRoomInviteRequest struct {
	Accept bool `json:"accept"`
}

type InviteUserRequest struct {
	InviteesId datatypes.JSON `json:"invitees" swaggertype:"array,string"`
}
