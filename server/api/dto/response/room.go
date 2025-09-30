package response

import (
	"time"

	"github.com/RowenTey/JustJio/server/api/model"
)

type AttendeesDto struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type RoomDto struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Time          string         `json:"time"`
	Venue         string         `json:"venue"`
	VenueUrl      string         `json:"venueUrl"`
	Date          time.Time      `json:"date"`
	Description   string         `json:"description"`
	Consolidated  string         `json:"consolidated"`
	IsClosed      bool           `json:"isClosed"`
	IsPrivate     bool           `json:"isPrivate"`
	ImageUrl      string         `json:"imageUrl"`
	Host          AttendeesDto   `json:"host"`
	NoOfAttendees int            `json:"noOfAttendees"`
	Attendees     []AttendeesDto `json:"attendees"`
}

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

type GetRoomResponse struct {
	Room RoomDto `json:"room"`
}
