package response

import (
	"time"
)

type AttendeesDto struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Picture  string `json:"pictureUrl"`
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

type RoomListDto struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	IsClosed      bool         `json:"isClosed"`
	IsPrivate     bool         `json:"isPrivate"`
	ImageUrl      string       `json:"imageUrl"`
	Host          AttendeesDto `json:"host"`
	NoOfAttendees int          `json:"noOfAttendees"`
}

type SimplifiedRoomDto struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Time          string    `json:"time"`
	Venue         string    `json:"venue"`
	VenueUrl      string    `json:"venueUrl"`
	Date          time.Time `json:"date"`
	Description   string    `json:"description"`
	ImageUrl      string    `json:"imageUrl"`
	IsPrivate     bool      `json:"isPrivate"`
	NoOfAttendees int       `json:"noOfAttendees"`
}

type RoomInviteDto struct {
	ID        uint              `json:"id"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	User      AttendeesDto      `json:"user"`
	Inviter   AttendeesDto      `json:"inviter"`
	Room      SimplifiedRoomDto `json:"room"`
}

type CountResponse struct {
	Count int `json:"count"`
}
