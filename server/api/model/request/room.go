package request

import (
	"time"

	"github.com/RowenTey/JustJio/server/api/model"

	"gorm.io/datatypes"
)

type CreateRoomRequest struct {
	Room       model.Room     `json:"room"`
	InviteesId datatypes.JSON `json:"invitees" swaggertype:"array,string"`
}

type RespondToRoomInviteRequest struct {
	Accept bool `json:"accept"`
}

type InviteUserRequest struct {
	InviteesId datatypes.JSON `json:"invitees" swaggertype:"array,string"`
}

type UpdateRoomRequest struct {
	Venue       string    `json:"venue"`
	PlaceId     string    `json:"placeId"`
	Date        time.Time `json:"date"`
	Time        string    `json:"time"`
	Description string    `json:"description"`
}
