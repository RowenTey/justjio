package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	modelLocation "github.com/RowenTey/JustJio/server/api/model/location"
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
	billRepo         repository.BillRepository
	httpClient       utils.HTTPClient
	googleMapsApiKey string
	logger           *logrus.Entry
}

func NewRoomService(
	db *gorm.DB,
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
	billRepo repository.BillRepository,
	httpClient utils.HTTPClient,
	googleMapsApiKey string,
	logger *logrus.Logger,
) *RoomService {
	return &RoomService{
		db:               db,
		roomRepo:         roomRepo,
		userRepo:         userRepo,
		billRepo:         billRepo,
		googleMapsApiKey: googleMapsApiKey,
		httpClient:       httpClient,
		logger:           utils.AddServiceField(logger, "RoomService"),
	}
}

func (rs *RoomService) CreateRoomWithInvites(
	room *model.Room, userId, placeId string, inviteesIds *[]uint) (*model.Room, *[]model.RoomInvite, error) {
	var invites []model.RoomInvite

	if err := database.RunInTransaction(rs.db, func(tx *gorm.DB) error {
		userRepoTx := rs.userRepo.WithTx(tx)
		roomRepoTx := rs.roomRepo.WithTx(tx)

		host, err := userRepoTx.FindByID(userId)
		if err != nil {
			return err
		}

		invitees, err := userRepoTx.FindByIDs(inviteesIds)
		if err != nil {
			return err
		}

		// Fetch the Google Maps URI
		googleMapsUri, err := rs.fetchGoogleMapsUri(placeId)
		if err != nil {
			return err
		}

		room.VenueUrl = googleMapsUri
		room.HostID = host.ID
		room.Users = append(room.Users, *host)
		if err := roomRepoTx.Create(room); err != nil {
			return err
		}

		for _, user := range *invitees {
			invite := model.RoomInvite{
				RoomID:    room.ID,
				UserID:    user.ID,
				InviterID: host.ID,
				Status:    "pending",
			}
			invites = append(invites, invite)
		}
		if err := roomRepoTx.CreateInvites(&invites); err != nil {
			return err
		}

		return err
	}); err != nil {
		return nil, nil, err
	}

	rs.logger.Info("Created room with ID: ", room.ID)
	rs.logger.Infof("Invited users: %v", invites)

	return room, &invites, nil
}

func (rs *RoomService) GetRooms(userId string, page int) (*[]model.Room, error) {
	return rs.roomRepo.GetUserRooms(userId, page, ROOM_PAGE_SIZE)
}

func (rs *RoomService) GetNumRooms(userId string) (int64, error) {
	return rs.roomRepo.CountUserRooms(userId)
}

func (rs *RoomService) GetUnjoinedPublicRooms(userId string) (*[]model.Room, error) {
	return rs.roomRepo.GetUnjoinedRoomsByIsPrivate(userId, false)
}

func (rs *RoomService) GetRoomById(roomId string) (*model.Room, error) {
	return rs.roomRepo.GetByID(roomId)
}

func (rs *RoomService) GetRoomInvites(userId string) (*[]model.RoomInvite, error) {
	return rs.roomRepo.GetPendingInvites(userId)
}

func (rs *RoomService) GetNumRoomInvites(userId string) (int64, error) {
	return rs.roomRepo.CountPendingInvites(userId)
}

func (rs *RoomService) GetRoomAttendees(roomId string) (*[]model.User, error) {
	return rs.roomRepo.GetRoomAttendees(roomId)
}

func (rs *RoomService) GetRoomAttendeesIds(roomId string) (*[]string, error) {
	return rs.roomRepo.GetRoomAttendeeIDs(roomId)
}

func (rs *RoomService) CloseRoom(roomId string, userId string) error {
	return database.RunInTransaction(rs.db, func(tx *gorm.DB) error {
		roomRepoTx := rs.roomRepo.WithTx(tx)
		billRepoTx := rs.billRepo.WithTx(tx)

		if status, err := billRepoTx.GetRoomBillConsolidationStatus(roomId); err != nil {
			return err
		} else if status == repository.UNCONSOLIDATED {
			return ErrRoomHasUnconsolidatedBills
		}

		room, err := roomRepoTx.GetByID(roomId)
		if err != nil {
			return err
		}

		if utils.UIntToString(room.HostID) != userId {
			return ErrInvalidHost
		}

		room.IsClosed = true
		if err := roomRepoTx.UpdateRoom(room); err != nil {
			return err
		}

		// TODO: Should we delete the room invites?
		return roomRepoTx.DeletePendingInvites(roomId)
	})
}

func (rs *RoomService) UpdateRoomInviteStatus(roomId string, userId string, status string) error {
	if status != "accepted" && status != "rejected" {
		return ErrInvalidRoomStatus
	}

	return database.RunInTransaction(rs.db, func(tx *gorm.DB) error {
		roomRepoTx := rs.roomRepo.WithTx(tx)
		userRepoTx := rs.userRepo.WithTx(tx)

		// Update the invite status
		if err := roomRepoTx.UpdateInviteStatus(roomId, userId, status); err != nil {
			return err
		}

		// If the invite is rejected, we don't need to update the room
		if status == "rejected" {
			return nil
		}

		room, err := roomRepoTx.GetByID(roomId)
		if err != nil {
			return err
		}

		user, err := userRepoTx.FindByID(userId)
		if err != nil {
			return err
		}

		// Update room info
		room.AttendeesCount++
		room.Users = append(room.Users, *user)
		return roomRepoTx.UpdateRoom(room)
	})
}

func (rs *RoomService) JoinRoom(roomId, userId string) (*model.Room, *[]model.User, error) {
	// Check if user is already in room
	if inRoom, err := rs.roomRepo.IsUserInRoom(roomId, userId); err != nil {
		return nil, nil, err
	} else if inRoom {
		return nil, nil, ErrAlreadyInRoom
	}

	// TODO: check if user is invited if room is private

	room, err := rs.roomRepo.GetByID(roomId)
	if err != nil {
		return nil, nil, err
	}

	user, err := rs.userRepo.FindByID(userId)
	if err != nil {
		return nil, nil, err
	}

	// Update room info
	room.AttendeesCount++
	room.Users = append(room.Users, *user)
	err = rs.roomRepo.UpdateRoom(room)
	if err != nil {
		return nil, nil, err
	}

	attendees, err := rs.GetRoomAttendees(roomId)
	if err != nil {
		return nil, nil, err
	}

	return room, attendees, nil
}

func (rs *RoomService) RespondToRoomInvite(
	roomId string,
	userId string,
	accept bool,
) (*model.Room, *[]model.User, error) {
	status := "accepted"
	if !accept {
		status = "rejected"
	}

	err := rs.UpdateRoomInviteStatus(roomId, userId, status)
	if err != nil {
		return nil, nil, err
	}

	// No room or attendees to return if invite is rejected
	if !accept {
		return nil, nil, nil
	}

	room, err := rs.roomRepo.GetByID(roomId)
	if err != nil {
		return nil, nil, err
	}

	attendees, err := rs.GetRoomAttendees(roomId)
	if err != nil {
		return nil, nil, err
	}

	return room, attendees, nil
}

func (rs *RoomService) ValidateInvites(
	room *model.Room,
	inviter *model.User,
	users *[]model.User,
) error {
	rs.logger.Infof("Inviting users (%v) to room %s", users, room.ID)

	// TODO: Can optimize this
	// Check if users are already in room or have pending invites
	for _, user := range *users {
		// Check if user is already in room
		if inRoom, err := rs.roomRepo.IsUserInRoom(
			room.ID,
			strconv.FormatUint(uint64(user.ID), 10),
		); err != nil {
			return err
		} else if inRoom {
			return ErrAlreadyInRoom
		}

		// Check if user has already pending invite
		if hasInvite, err := rs.roomRepo.HasPendingInvites(
			room.ID,
			strconv.FormatUint(uint64(user.ID), 10),
		); err != nil {
			return err
		} else if hasInvite {
			return ErrAlreadyInvited
		}
	}

	return nil
}

func (rs *RoomService) InviteUsersToRoom(
	roomId string, inviterId string, inviteesIds *[]uint) (*[]model.RoomInvite, error) {
	var roomInvites []model.RoomInvite

	err := database.RunInTransaction(rs.db, func(tx *gorm.DB) error {
		roomRepoTx := rs.roomRepo.WithTx(tx)
		userRepoTx := rs.userRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(roomId)
		if err != nil {
			return err
		}

		if utils.UIntToString(room.HostID) != inviterId {
			return ErrInvalidHost
		}

		inviter, err := userRepoTx.FindByID(inviterId)
		if err != nil {
			return err
		}

		invitees, err := userRepoTx.FindByIDs(inviteesIds)
		if err != nil {
			return err
		}

		if err := rs.ValidateInvites(room, inviter, invitees); err != nil {
			return err
		}

		for _, invitee := range *invitees {
			roomInvite := model.RoomInvite{
				RoomID:    room.ID,
				UserID:    invitee.ID,
				InviterID: inviter.ID,
				Status:    "pending",
			}
			roomInvites = append(roomInvites, roomInvite)
		}
		return rs.roomRepo.CreateInvites(&roomInvites)
	})

	return &roomInvites, err
}

func (rs *RoomService) LeaveRoom(roomId string, userId string) error {
	return database.RunInTransaction(rs.db, func(tx *gorm.DB) error {
		roomRepoTx := rs.roomRepo.WithTx(tx)
		billRepoTx := rs.billRepo.WithTx(tx)

		// TODO: check if user is involved in any bills first
		if status, err := billRepoTx.GetRoomBillConsolidationStatus(roomId); err != nil {
			return err
		} else if status == repository.UNCONSOLIDATED {
			return ErrRoomHasUnconsolidatedBills
		}

		room, err := roomRepoTx.GetByID(roomId)
		if err != nil {
			return err
		}

		if utils.UIntToString(room.HostID) == userId {
			return ErrLeaveRoomAsHost
		}

		// TODO: Should we delete the room invites?
		return roomRepoTx.RemoveUserFromRoom(roomId, userId)
	})
}

func (rs *RoomService) RemoveUserFromRoom(roomId string, userId string) error {
	return rs.roomRepo.RemoveUserFromRoom(roomId, userId)
}

func (rs *RoomService) GetUninvitedFriendsForRoom(roomId string, userId string) (*[]model.User, error) {
	return rs.userRepo.GetUninvitedFriends(roomId, userId)
}

func (rs *RoomService) QueryVenue(query string) (*[]modelLocation.Venue, error) {
	if query == "" {
		return nil, errors.New("location query cannot be empty")
	}

	requestBody := map[string]interface{}{
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	suggestions, _ := response["suggestions"].([]interface{})

	// Extract the predictions
	var predictions []modelLocation.Venue
	for _, suggestion := range suggestions {
		suggestionMap, ok := suggestion.(map[string]interface{})
		if !ok {
			continue
		}

		// Check for place prediction
		placePrediction, exists := suggestionMap["placePrediction"].(map[string]interface{})
		if !exists {
			continue
		}

		var venue modelLocation.Venue
		if text, exists := placePrediction["text"].(map[string]interface{}); exists {
			if address, ok := text["text"].(string); ok {
				venue.Address = address
			}
		}

		if placeId, ok := placePrediction["placeId"].(string); ok {
			venue.GoogleMapsPlaceID = placeId
		}

		if structuredFormat, exists := placePrediction["structuredFormat"].(map[string]interface{}); exists {
			if mainText, exists := structuredFormat["mainText"].(map[string]interface{}); exists {
				if name, ok := mainText["text"].(string); ok {
					venue.Name = name
				}
			}
		}

		predictions = append(predictions, venue)
	}

	return &predictions, nil
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var placeResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&placeResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	googleMapsUri, ok := placeResponse["googleMapsUri"].(string)
	if !ok {
		return "", errors.New("googleMapsUri not found in response")
	}

	return googleMapsUri, nil
}
