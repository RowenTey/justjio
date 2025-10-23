package response

import (
	"time"
)

type RoomDto struct {
	ID            string           `json:"id" binding:"required"`
	Name          string           `json:"name" binding:"required"`
	Time          string           `json:"time" binding:"required"`
	Venue         string           `json:"venue" binding:"required"`
	VenueUrl      string           `json:"venueUrl" binding:"required"`
	Date          time.Time        `json:"date" binding:"required"`
	Description   string           `json:"description" binding:"required"`
	Consolidated  string           `json:"consolidated" binding:"required"`
	IsClosed      bool             `json:"isClosed" binding:"required"`
	IsPrivate     bool             `json:"isPrivate" binding:"required"`
	ImageUrl      string           `json:"imageUrl" binding:"required"`
	Host          MinimalUserDto   `json:"host" binding:"required"`
	NoOfAttendees int              `json:"noOfAttendees" binding:"required"`
	Attendees     []MinimalUserDto `json:"attendees" binding:"required"`
}

type RoomListDto struct {
	ID            string         `json:"id" binding:"required"`
	Name          string         `json:"name" binding:"required"`
	IsClosed      bool           `json:"isClosed" binding:"required"`
	IsPrivate     bool           `json:"isPrivate" binding:"required"`
	ImageUrl      string         `json:"imageUrl" binding:"required"`
	Host          MinimalUserDto `json:"host" binding:"required"`
	NoOfAttendees int            `json:"noOfAttendees" binding:"required"`
}

type SimplifiedRoomDto struct {
	ID            string         `json:"id" binding:"required"`
	Name          string         `json:"name" binding:"required"`
	Time          string         `json:"time" binding:"required"`
	Venue         string         `json:"venue" binding:"required"`
	VenueUrl      string         `json:"venueUrl" binding:"required"`
	Date          time.Time      `json:"date" binding:"required"`
	Description   string         `json:"description" binding:"required"`
	ImageUrl      string         `json:"imageUrl" binding:"required"`
	IsPrivate     bool           `json:"isPrivate" binding:"required"`
	NoOfAttendees int            `json:"noOfAttendees" binding:"required"`
	Host          MinimalUserDto `json:"host" binding:"required"`
}

type RoomInviteDto struct {
	ID        uint              `json:"id" binding:"required"`
	Status    string            `json:"status" binding:"required"`
	CreatedAt time.Time         `json:"createdAt" binding:"required"`
	User      MinimalUserDto    `json:"user" binding:"required"`
	Inviter   MinimalUserDto    `json:"inviter" binding:"required"`
	Room      SimplifiedRoomDto `json:"room" binding:"required"`
}
