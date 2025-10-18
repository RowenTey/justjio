package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/RowenTey/JustJio/server/api/database"
	modelLocation "github.com/RowenTey/JustJio/server/api/dto/location"
	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/dto/response"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	ROOM_PAGE_SIZE = 6
)

var (
	ErrRoomHasUnconsolidatedBills = errors.New("cannot perform action with unconsolidated bills")
	ErrLeaveRoomAsHost            = errors.New("cannot leave room as host")
	ErrInvalidHost                = errors.New("user is not the host of the room")
	ErrInvalidRoomStatus          = errors.New("invalid room status")
	ErrAlreadyInRoom              = errors.New("user is already in room")
	ErrAlreadyInvited             = errors.New("user already has pending invite")
)

type RoomService struct {
	db               *gorm.DB
	roomRepo         repository.RoomRepository
	userRepo         repository.UserRepository
	httpClient       utils.HTTPClient
	googleMapsApiKey string
	logger           *logrus.Entry
}

func NewRoomService(
	db *gorm.DB,
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
	httpClient utils.HTTPClient,
	googleMapsApiKey string,
	logger *logrus.Logger,
) *RoomService {
	return &RoomService{
		db:               db,
		roomRepo:         roomRepo,
		userRepo:         userRepo,
		googleMapsApiKey: googleMapsApiKey,
		httpClient:       httpClient,
		logger:           utils.AddServiceField(logger, "RoomService"),
	}
}

func (rs *RoomService) CreateRoomWithInvites(
	ctx context.Context,
	room *model.Room,
	userId string,
	inviteesIds []string,
) (string, error) {
	var createdRoomId string

	if err := database.RunInTransaction(rs.db, sql.LevelRepeatableRead, func(tx *gorm.DB) error {
		var invites []model.RoomInvite

		userRepoTx := rs.userRepo.WithTx(tx)
		roomRepoTx := rs.roomRepo.WithTx(tx)

		host, err := userRepoTx.FindByID(ctx, userId)
		if err != nil {
			return err
		}

		invitees, err := userRepoTx.FindByIDs(ctx, inviteesIds)
		if err != nil {
			return err
		}

		// Fetch the Google Maps URI
		googleMapsUri, err := rs.fetchGoogleMapsUri(room.VenuePlaceId)
		if err != nil {
			return err
		}

		room.NoOfAttendees = 1
		room.VenueUrl = googleMapsUri
		room.HostID = host.ID
		room.Users = append(room.Users, *host)
		if err := roomRepoTx.Create(ctx, room); err != nil {
			return err
		}

		for _, user := range invitees {
			invite := model.RoomInvite{
				RoomID:    room.ID,
				UserID:    user.ID,
				InviterID: host.ID,
				Status:    "pending",
			}
			invites = append(invites, invite)
		}
		if err := roomRepoTx.CreateInvites(ctx, invites); err != nil {
			return err
		}

		host.NoOfRooms++
		if err := userRepoTx.Update(ctx, host); err != nil {
			return err
		}

		if err := userRepoTx.UpdateNoOfPendingRoomInvites(ctx, inviteesIds, 1); err != nil {
			return err
		}

		rs.logger.Info("Room ", room.Name, " created with ID: ", room.ID)
		createdRoomId = room.ID

		return err
	}); err != nil {
		return "", err
	}

	return createdRoomId, nil
}

func (rs *RoomService) GetRooms(ctx context.Context, userId string, page int) ([]response.RoomListDto, error) {
	rooms, err := rs.roomRepo.GetUserRooms(ctx, userId, page, ROOM_PAGE_SIZE)
	if err != nil {
		return nil, err
	}

	roomDtos := make([]response.RoomListDto, len(rooms))
	for i, room := range rooms {
		roomDtos[i] = response.RoomListDto{
			ID:            room.ID,
			Name:          room.Name,
			IsClosed:      room.IsClosed,
			IsPrivate:     room.IsPrivate,
			ImageUrl:      room.ImageUrl,
			NoOfAttendees: room.NoOfAttendees,
			Host: response.AttendeesDto{
				ID:       room.Host.ID,
				Username: room.Host.Username,
				Picture:  room.Host.PictureUrl,
			},
		}
	}

	return roomDtos, nil
}

func (rs *RoomService) GetNumRooms(ctx context.Context, userId string) (int64, error) {
	return rs.roomRepo.CountUserRooms(ctx, userId)
}

func (rs *RoomService) GetUnjoinedPublicRooms(ctx context.Context, userId string) ([]response.RoomListDto, error) {
	rooms, err := rs.roomRepo.GetUnjoinedRoomsByIsPrivate(ctx, userId, false)
	if err != nil {
		return nil, err
	}

	roomDtos := make([]response.RoomListDto, len(rooms))
	for i, room := range rooms {
		roomDtos[i] = response.RoomListDto{
			ID:            room.ID,
			Name:          room.Name,
			IsClosed:      room.IsClosed,
			IsPrivate:     room.IsPrivate,
			ImageUrl:      room.ImageUrl,
			NoOfAttendees: room.NoOfAttendees,
			Host: response.AttendeesDto{
				ID:       room.Host.ID,
				Username: room.Host.Username,
				Picture:  room.Host.PictureUrl,
			},
		}
	}

	return roomDtos, nil
}

func (rs *RoomService) GetRoomAttendeesIds(ctx context.Context, roomId string) ([]string, error) {
	return rs.roomRepo.GetRoomAttendeeIDs(ctx, roomId)
}

func (rs *RoomService) GetRoomById(ctx context.Context, roomId string) (*response.RoomDto, error) {
	room, err := rs.roomRepo.GetByIDWithAttendees(ctx, roomId)
	if err != nil {
		return nil, err
	}

	dto := &response.RoomDto{
		ID:           room.ID,
		Name:         room.Name,
		Time:         room.Time,
		Venue:        room.Venue,
		VenueUrl:     room.VenueUrl,
		Date:         room.Date,
		Description:  room.Description,
		Consolidated: room.Consolidated,
		IsClosed:     room.IsClosed,
		IsPrivate:    room.IsPrivate,
		ImageUrl:     room.ImageUrl,
		Host: response.AttendeesDto{
			ID:       room.Host.ID,
			Username: room.Host.Username,
			Picture:  room.Host.PictureUrl,
		},
		NoOfAttendees: room.NoOfAttendees,
		Attendees:     []response.AttendeesDto{},
	}

	for _, u := range room.Users {
		dto.Attendees = append(dto.Attendees, response.AttendeesDto{
			ID:       u.ID,
			Username: u.Username,
			Picture:  u.PictureUrl,
		})
	}

	return dto, nil
}

func (rs *RoomService) GetRoomInvites(ctx context.Context, userId string) ([]response.RoomInviteDto, error) {
	invites, err := rs.roomRepo.GetPendingInvites(ctx, userId)
	if err != nil {
		return nil, err
	}

	roomInviteDtos := make([]response.RoomInviteDto, len(invites))
	for i, invite := range invites {
		roomInviteDtos[i] = response.RoomInviteDto{
			ID:        invite.ID,
			Status:    invite.Status,
			CreatedAt: invite.CreatedAt,
			User: response.AttendeesDto{
				ID:       invite.User.ID,
				Username: invite.User.Username,
				Picture:  invite.User.PictureUrl,
			},
			Inviter: response.AttendeesDto{
				ID:       invite.Inviter.ID,
				Username: invite.Inviter.Username,
				Picture:  invite.Inviter.PictureUrl,
			},
			Room: response.SimplifiedRoomDto{
				ID:            invite.Room.ID,
				Name:          invite.Room.Name,
				Time:          invite.Room.Time,
				Venue:         invite.Room.Venue,
				VenueUrl:      invite.Room.VenueUrl,
				Date:          invite.Room.Date,
				Description:   invite.Room.Description,
				ImageUrl:      invite.Room.ImageUrl,
				IsPrivate:     invite.Room.IsPrivate,
				NoOfAttendees: invite.Room.NoOfAttendees,
			},
		}
	}

	return roomInviteDtos, nil
}

func (rs *RoomService) GetNumRoomInvites(ctx context.Context, userId string) (int64, error) {
	return rs.roomRepo.CountPendingInvites(ctx, userId)
}

func (rs *RoomService) UpdateRoom(
	ctx context.Context,
	updateReq *request.EditRoomRequest,
	roomId,
	userId string,
) error {
	room, err := rs.roomRepo.GetByID(ctx, roomId)
	if err != nil {
		return err
	}

	if utils.UIntToString(room.HostID) != userId {
		return ErrInvalidHost
	}

	if updateReq.Name != nil {
		room.Name = *updateReq.Name
	}

	if updateReq.Time != nil {
		room.Time = *updateReq.Time
	}

	if updateReq.Date != nil {
		room.Date = *updateReq.Date
	}

	if updateReq.Description != nil {
		room.Description = *updateReq.Description
	}

	if updateReq.ImageUrl != nil {
		room.ImageUrl = *updateReq.ImageUrl
	}

	// TODO: Check if this is the desired logic
	if updateReq.VenuePlaceId != nil && room.VenuePlaceId != *updateReq.VenuePlaceId {
		room.VenuePlaceId = *updateReq.VenuePlaceId

		if updateReq.Venue != nil {
			room.Venue = *updateReq.Venue
		}

		room.VenueUrl, err = rs.fetchGoogleMapsUri(*updateReq.VenuePlaceId)
		if err != nil {
			return fmt.Errorf("failed to fetch Google Maps URI: %v", err)
		}
	}

	if err := rs.roomRepo.Update(ctx, room); err != nil {
		return fmt.Errorf("failed to update room: %v", err)
	}

	rs.logger.Info("Room " + room.Name + " edited successfully.")
	return nil
}

func (rs *RoomService) CloseRoom(ctx context.Context, roomId string, userId string) error {
	return database.RunInTransaction(rs.db, sql.LevelDefault, func(tx *gorm.DB) error {
		roomRepoTx := rs.roomRepo.WithTx(tx)
		userRepoTx := rs.userRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(ctx, roomId)
		if err != nil {
			return err
		}

		if utils.UIntToString(room.HostID) != userId {
			return ErrInvalidHost
		}

		if room.Consolidated == "UNCONSOLIDATED" {
			return ErrRoomHasUnconsolidatedBills
		}

		room.IsClosed = true
		if err := roomRepoTx.Update(ctx, room); err != nil {
			return err
		}

		inviteesId, err := roomRepoTx.GetPendingInviteUsers(ctx, roomId)
		if err != nil {
			return err
		}

		// Decrement number of pending invites for each invitee who haven't responded
		if err := userRepoTx.UpdateNoOfPendingRoomInvites(ctx, inviteesId, -1); err != nil {
			return err
		}

		return roomRepoTx.DeletePendingInvites(ctx, roomId)
	})
}

func (rs *RoomService) JoinRoom(ctx context.Context, roomId, userId string) (*response.RoomDto, error) {
	var room *model.Room

	if err := database.RunInTransaction(rs.db, sql.LevelRepeatableRead, func(tx *gorm.DB) error {
		// Check if user is already in room
		if inRoom, err := rs.roomRepo.IsUserInRoom(ctx, roomId, userId); err != nil {
			return err
		} else if inRoom {
			return ErrAlreadyInRoom
		}

		// TODO: check if user is invited if room is private
		var err error
		room, err = rs.roomRepo.GetByIDWithAttendees(ctx, roomId)
		if err != nil {
			return err
		}

		user, err := rs.userRepo.FindByID(ctx, userId)
		if err != nil {
			return err
		}

		user.NoOfRooms++
		if err := rs.userRepo.Update(ctx, user); err != nil {
			return err
		}

		room.NoOfAttendees++
		room.Users = append(room.Users, *user)
		if err = rs.roomRepo.Update(ctx, room); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	dto := &response.RoomDto{
		ID:           room.ID,
		Name:         room.Name,
		Time:         room.Time,
		Venue:        room.Venue,
		VenueUrl:     room.VenueUrl,
		Date:         room.Date,
		Description:  room.Description,
		Consolidated: room.Consolidated,
		IsClosed:     room.IsClosed,
		IsPrivate:    room.IsPrivate,
		ImageUrl:     room.ImageUrl,
		Host: response.AttendeesDto{
			ID:       room.Host.ID,
			Username: room.Host.Username,
		},
		NoOfAttendees: room.NoOfAttendees,
		Attendees:     []response.AttendeesDto{},
	}

	for _, u := range room.Users {
		dto.Attendees = append(dto.Attendees, response.AttendeesDto{
			ID:       u.ID,
			Username: u.Username,
		})
	}

	return dto, nil
}

func (rs *RoomService) RespondToRoomInvite(
	ctx context.Context,
	roomId string,
	userId string,
	accept bool,
) (*response.RoomDto, error) {
	status := "accepted"
	if !accept {
		status = "rejected"
	}

	room, err := rs.updateRoomInviteStatus(ctx, roomId, userId, status)
	if err != nil {
		return nil, err
	}

	// No room or attendees to return if invite is rejected
	if !accept {
		return nil, err
	}

	dto := &response.RoomDto{
		ID:           room.ID,
		Name:         room.Name,
		Time:         room.Time,
		Venue:        room.Venue,
		VenueUrl:     room.VenueUrl,
		Date:         room.Date,
		Description:  room.Description,
		Consolidated: room.Consolidated,
		IsClosed:     room.IsClosed,
		IsPrivate:    room.IsPrivate,
		ImageUrl:     room.ImageUrl,
		Host: response.AttendeesDto{
			ID:       room.Host.ID,
			Username: room.Host.Username,
		},
		NoOfAttendees: room.NoOfAttendees,
		Attendees:     []response.AttendeesDto{},
	}

	return dto, nil
}

func (rs *RoomService) InviteUsersToRoom(
	ctx context.Context,
	roomId string,
	inviterId string,
	inviteesIds []string,
) ([]model.RoomInvite, error) {
	var roomInvites []model.RoomInvite

	err := database.RunInTransaction(rs.db, sql.LevelRepeatableRead, func(tx *gorm.DB) error {
		roomRepoTx := rs.roomRepo.WithTx(tx)
		userRepoTx := rs.userRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(ctx, roomId)
		if err != nil {
			return err
		}

		if utils.UIntToString(room.HostID) != inviterId {
			return ErrInvalidHost
		}

		inviter, err := userRepoTx.FindByID(ctx, inviterId)
		if err != nil {
			return err
		}

		invitees, err := userRepoTx.FindByIDs(ctx, inviteesIds)
		if err != nil {
			return err
		}

		if err := rs.validateInvites(ctx, room, inviteesIds); err != nil {
			return err
		}

		if err := userRepoTx.UpdateNoOfPendingRoomInvites(ctx, inviteesIds, 1); err != nil {
			return err
		}

		for _, invitee := range invitees {
			roomInvite := model.RoomInvite{
				RoomID:    room.ID,
				UserID:    invitee.ID,
				InviterID: inviter.ID,
				Status:    "pending",
			}
			roomInvites = append(roomInvites, roomInvite)
		}

		return roomRepoTx.CreateInvites(ctx, roomInvites)
	})

	return roomInvites, err
}

func (rs *RoomService) LeaveRoom(ctx context.Context, roomId string, userId string) error {
	return database.RunInTransaction(rs.db, sql.LevelDefault, func(tx *gorm.DB) error {
		roomRepoTx := rs.roomRepo.WithTx(tx)
		userRepoTx := rs.userRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(ctx, roomId)
		if err != nil {
			return err
		}

		// TODO: check if user is involved in any bills first
		if room.Consolidated == "UNCONSOLIDATED" {
			return ErrRoomHasUnconsolidatedBills
		}

		user, err := userRepoTx.FindByID(ctx, userId)
		if err != nil {
			return err
		}

		if utils.UIntToString(room.HostID) == userId {
			return ErrLeaveRoomAsHost
		}

		user.NoOfRooms--
		if err := userRepoTx.Update(ctx, user); err != nil {
			return err
		}

		room.NoOfAttendees--
		if err := roomRepoTx.Update(ctx, room); err != nil {
			return err
		}

		return roomRepoTx.RemoveUserFromRoom(ctx, roomId, userId)
	})
}

func (rs *RoomService) GetUninvitedFriendsForRoom(ctx context.Context, roomId string, userId string) ([]response.AttendeesDto, error) {
	friends, err := rs.userRepo.GetUninvitedFriends(ctx, roomId, userId)
	if err != nil {
		return nil, err
	}

	friendDtos := make([]response.AttendeesDto, len(friends))
	for i, friend := range friends {
		friendDtos[i] = response.AttendeesDto{
			ID:       friend.ID,
			Username: friend.Username,
			Picture:  friend.PictureUrl,
		}
	}

	return friendDtos, nil
}

func (rs *RoomService) QueryVenue(query string) ([]modelLocation.Venue, error) {
	if query == "" {
		return nil, errors.New("location query cannot be empty")
	}

	requestBody := map[string]any{
		"input":               query,
		"includedRegionCodes": []string{"sg", "my"},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Returns up to 5 predictions
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		"https://places.googleapis.com/v1/places:autocomplete",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set custom headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", rs.googleMapsApiKey)
	req.Header.Set("X-Goog-FieldMask", "suggestions.placePrediction.text.text,suggestions.placePrediction.placeId,suggestions.placePrediction.structuredFormat.mainText.text")

	resp, err := rs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			rs.logger.Error("Failed to close response body: ", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	suggestions, _ := response["suggestions"].([]any)

	// Extract the predictions
	var predictions []modelLocation.Venue
	for _, suggestion := range suggestions {
		suggestionMap, ok := suggestion.(map[string]any)
		if !ok {
			continue
		}

		// Check for place prediction
		placePrediction, exists := suggestionMap["placePrediction"].(map[string]any)
		if !exists {
			continue
		}

		var venue modelLocation.Venue
		if text, exists := placePrediction["text"].(map[string]any); exists {
			if address, ok := text["text"].(string); ok {
				venue.Address = address
			}
		}

		if placeId, ok := placePrediction["placeId"].(string); ok {
			venue.GoogleMapsPlaceID = placeId
		}

		if structuredFormat, exists := placePrediction["structuredFormat"].(map[string]any); exists {
			if mainText, exists := structuredFormat["mainText"].(map[string]any); exists {
				if name, ok := mainText["text"].(string); ok {
					venue.Name = name
				}
			}
		}

		predictions = append(predictions, venue)
	}

	return predictions, nil
}

func (rs *RoomService) updateRoomInviteStatus(
	ctx context.Context,
	roomId string,
	userId string,
	status string,
) (*model.Room, error) {
	if status != "accepted" && status != "rejected" {
		return nil, ErrInvalidRoomStatus
	}

	var room *model.Room
	if err := database.RunInTransaction(rs.db, sql.LevelRepeatableRead, func(tx *gorm.DB) error {
		roomRepoTx := rs.roomRepo.WithTx(tx)
		userRepoTx := rs.userRepo.WithTx(tx)

		// Update the invite status
		if err := roomRepoTx.UpdateInviteStatus(ctx, roomId, userId, status); err != nil {
			return err
		}

		user, err := userRepoTx.FindByID(ctx, userId)
		if err != nil {
			return err
		}

		user.NoOfPendingRoomInvites--
		if err := userRepoTx.Update(ctx, user); err != nil {
			return err
		}

		// If the invite is rejected, we don't need to update the room
		if status == "rejected" {
			return nil
		}

		room, err = roomRepoTx.GetByIDWithAttendees(ctx, roomId)
		if err != nil {
			return err
		}

		user.NoOfRooms++
		if err = userRepoTx.Update(ctx, user); err != nil {
			return err
		}

		room.NoOfAttendees++
		room.Users = append(room.Users, *user)
		return roomRepoTx.Update(ctx, room)
	}); err != nil {
		return nil, err
	}

	return room, nil
}

func (rs *RoomService) validateInvites(
	ctx context.Context,
	room *model.Room,
	userIds []string,
) error {
	rs.logger.Infof("Inviting users (%v) to room %s", userIds, room.ID)

	attendees, err := rs.roomRepo.GetRoomAttendeeIDs(ctx, room.ID)
	if err != nil {
		return err
	}

	invitees, err := rs.roomRepo.GetPendingInviteUsers(ctx, room.ID)
	if err != nil {
		return err
	}

	attendeeSet := make(map[string]bool)
	for _, attendeeID := range attendees {
		attendeeSet[attendeeID] = true
	}

	inviteeSet := make(map[string]bool)
	for _, inviteeID := range invitees {
		inviteeSet[inviteeID] = true
	}

	// Check if users are already in room or have pending invites
	for _, userID := range userIds {
		if attendeeSet[userID] {
			return ErrAlreadyInRoom
		}
		if inviteeSet[userID] {
			return ErrAlreadyInvited
		}
	}

	return nil
}

func (rs *RoomService) fetchGoogleMapsUri(placeId string) (string, error) {
	reqUrl := fmt.Sprintf("https://places.googleapis.com/v1/places/%s", placeId)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Goog-Api-Key", rs.googleMapsApiKey)
	req.Header.Set("X-Goog-FieldMask", "googleMapsUri")

	resp, err := rs.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			rs.logger.Error("Failed to close response body: ", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var placeResponse map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&placeResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	googleMapsUri, ok := placeResponse["googleMapsUri"].(string)
	if !ok {
		return "", errors.New("googleMapsUri not found in response")
	}

	return googleMapsUri, nil
}
