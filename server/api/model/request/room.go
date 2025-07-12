package request

import (
	"github.com/RowenTey/JustJio/server/api/model"

	"gorm.io/datatypes"
)

type CreateRoomRequest struct {
	Room       model.Room     `json:"room"`
	PlaceId    string         `json:"placeId"`
	InviteesId datatypes.JSON `json:"invitees" swaggertype:"array,string"`
}

type RespondToRoomInviteRequest struct {
	Accept bool `json:"accept"`
}

type InviteUserRequest struct {
	InviteesId datatypes.JSON `json:"invitees" swaggertype:"array,string"`
}
