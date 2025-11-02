package request

import (
	"time"
)

// CreateRoomRequest represents the request body for creating a new room
type CreateRoomRequest struct {
	Name         string    `json:"name" validate:"required,min=3,max=100" example:"John's Birthday Party"`
	Time         string    `json:"time" validate:"required" example:"7:00 PM"`
	Venue        string    `json:"venue" validate:"required" example:"Marina Bay Sands"`
	VenuePlaceId string    `json:"venuePlaceId" validate:"required" example:"ChIJkxHPFjMZ2jERPRhLUvKGfFk"`
	Date         time.Time `json:"date" validate:"required" example:"2025-12-25T19:00:00Z"`
	Description  string    `json:"description" validate:"max=500" example:"Let's celebrate John's birthday!"`
	IsPrivate    bool      `json:"isPrivate" example:"false"`
	ImageUrl     string    `json:"imageUrl" validate:"required,url" example:"https://example.com/party.jpg"`
	Invitees     []string  `json:"invitees" swaggertype:"array,string" example:"1,2,3"` // Array of user IDs
}

// EditRoomRequest represents the request body for editing an existing room
type EditRoomRequest struct {
	Name         *string    `json:"name,omitempty" validate:"omitempty,min=3,max=100" example:"Updated Party Name"`
	Time         *string    `json:"time,omitempty" example:"8:00 PM"`
	Venue        *string    `json:"venue,omitempty" example:"Sentosa Beach"`
	VenuePlaceId *string    `json:"venuePlaceId,omitempty" example:"ChIJkxHPFjMZ2jERPRhLUvKGfFk"`
	Date         *time.Time `json:"date,omitempty" example:"2025-12-26T19:00:00Z"`
	Description  *string    `json:"description,omitempty" validate:"omitempty,max=500" example:"Updated description"`
	ImageUrl     *string    `json:"imageUrl,omitempty" validate:"omitempty,url" example:"https://example.com/updated.jpg"`
}

// RespondToRoomInviteRequest represents the request body for accepting/rejecting room invites
type RespondToRoomInviteRequest struct {
	Accept bool `json:"accept" example:"true"`
}

// InviteUserRequest represents the request body for inviting users to a room
type InviteUserRequest struct {
	Invitees []string `json:"invitees" validate:"required,min=1" swaggertype:"array,string" example:"4,5,6"` // Array of user IDs
}
