package handlers

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/model/response"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type RoomHandler struct {
	roomService *services.RoomService
	logger      *log.Entry
}

func NewRoomHandler(
	roomService *services.RoomService,
	logger *log.Logger,
) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		logger:      utils.AddServiceField(logger, "RoomHandler"),
	}
}

func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	room, err := h.roomService.GetRoomById(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}
	return utils.HandleSuccess(c, "Retrieved room successfully", room)
}

func (h *RoomHandler) GetRooms(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	page := c.QueryInt("page", 1)

	rooms, err := h.roomService.GetRooms(userId, page)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No rooms found")
	}

	return utils.HandleSuccess(c, "Retrieved rooms successfully", rooms)
}

func (h *RoomHandler) GetNumRooms(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	numRooms, err := h.roomService.GetNumRooms(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No rooms found")
	}

	response := response.GetNumRoomsResponse{Count: int(numRooms)}
	return utils.HandleSuccess(c, "Retrieved number of rooms successfully", response)
}

func (h *RoomHandler) GetUnjoinedPublicRooms(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	rooms, err := h.roomService.GetUnjoinedPublicRooms(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No public rooms found")
	}

	return utils.HandleSuccess(c, "Retrieved public rooms successfully", rooms)
}

func (h *RoomHandler) GetRoomInvitations(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	invites, err := h.roomService.GetRoomInvites(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No room invitations found")
	}

	return utils.HandleSuccess(c, "Retrieved room invitations successfully", invites)
}

func (h *RoomHandler) GetNumRoomInvitations(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	numInvites, err := h.roomService.GetNumRoomInvites(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No room invitations found")
	}

	response := response.GetNumRoomInvitationsResponse{Count: int(numInvites)}
	return utils.HandleSuccess(c, "Retrieved number of invitations successfully", response)
}

func (h *RoomHandler) GetRoomAttendees(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	attendees, err := h.roomService.GetRoomAttendees(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No attendees found")
	}

	return utils.HandleSuccess(c, "Retrieved room attendees successfully", attendees)
}

func (h *RoomHandler) GetUninvitedFriendsForRoom(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	friends, err := h.roomService.GetUninvitedFriendsForRoom(roomId, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No uninvited friends found")
	}

	return utils.HandleSuccess(c, "Retrieved uninvited friends successfully", friends)
}

func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	var request request.CreateRoomRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	var inviteesIds []string
	if err := json.Unmarshal([]byte(request.InviteesId), &inviteesIds); err != nil {
		h.logger.Error("Failed to parse invitees IDs:", err)
		return utils.HandleInvalidInputError(c, err)
	}

	// convert to uint slice
	var inviteesIdsUint []uint
	for _, id := range inviteesIds {
		var uid uint
		if err := json.Unmarshal([]byte(id), &uid); err != nil {
			return utils.HandleInvalidInputError(c, err)
		}
		inviteesIdsUint = append(inviteesIdsUint, uid)
	}

	room, invites, err := h.roomService.CreateRoomWithInvites(
		&request.Room, userId, &inviteesIdsUint)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Failed to create room and invites")
	}

	response := response.CreateRoomResponse{
		Room:    *room,
		Invites: *invites,
	}
	h.logger.Info("Room " + room.Name + " created successfully.")
	return utils.HandleSuccess(c, "Created room successfully", response)
}

func (h *RoomHandler) EditRoom(c *fiber.Ctx) error {
	var request request.UpdateRoomRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	roomId := c.Params("roomId")
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	room, err := h.roomService.UpdateRoom(&request, roomId, userId)
	if err != nil {
		if errors.Is(err, services.ErrInvalidHost) {
			return utils.HandleError(c, fiber.StatusUnauthorized, "Only hosts can edit rooms", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, RoomNotFoundErrorMsg)
	}

	h.logger.Info("Room " + room.Name + " edited successfully.")
	return utils.HandleSuccess(c, "Edited room successfully", room)
}

func (h *RoomHandler) CloseRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	if err := h.roomService.CloseRoom(roomId, userId); err != nil {
		if errors.Is(err, services.ErrInvalidHost) {
			return utils.HandleError(
				c, fiber.StatusUnauthorized, "Only hosts are allowed to close rooms", err)
		} else if errors.Is(err, services.ErrRoomHasUnconsolidatedBills) {
			return utils.HandleError(
				c, fiber.StatusConflict, "Cannot close room with unconsolidated bills", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	return utils.HandleSuccess(c, "Closed room successfully", nil)
}

func (h *RoomHandler) JoinRoom(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	room, attendees, err := h.roomService.JoinRoom(roomId, userId)
	if err != nil {
		if errors.Is(err, services.ErrAlreadyInRoom) {
			return utils.HandleError(c, fiber.StatusConflict, "User is already in room", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	roomResponse := response.JoinRoomResponse{
		Room:      *room,
		Attendees: *attendees,
	}
	h.logger.Info("User " + utils.GetUserInfoFromToken(token, "username") + " joined Room " + roomId + " successfully.")
	return utils.HandleSuccess(c, "Joined room successfully", roomResponse)
}

func (h *RoomHandler) RespondToRoomInvite(c *fiber.Ctx) error {
	var request request.RespondToRoomInviteRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	room, attendees, err := h.roomService.RespondToRoomInvite(roomId, userId, request.Accept)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	if !request.Accept {
		return utils.HandleSuccess(c, "Rejected room invitation successfully", nil)
	}

	roomResponse := response.JoinRoomResponse{
		Room:      *room,
		Attendees: *attendees,
	}
	h.logger.Info(
		"User " + utils.GetUserInfoFromToken(token, "username") + " joined Room " + roomId + " successfully.")
	return utils.HandleSuccess(c, "Joined room successfully", roomResponse)
}

func (h *RoomHandler) InviteUser(c *fiber.Ctx) error {
	var request request.InviteUserRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	var inviteesIds []string
	if err := json.Unmarshal([]byte(request.InviteesId), &inviteesIds); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	// convert to uint slice
	var inviteesIdsUint []uint
	for _, id := range inviteesIds {
		var uid uint
		if err := json.Unmarshal([]byte(id), &uid); err != nil {
			return utils.HandleInvalidInputError(c, err)
		}
		inviteesIdsUint = append(inviteesIdsUint, uid)
	}

	roomInvites, err := h.roomService.InviteUsersToRoom(
		roomId, userId, &inviteesIdsUint)
	if err != nil {
		if errors.Is(err, services.ErrInvalidHost) {
			return utils.HandleError(c, fiber.StatusUnauthorized, "Only hosts are allowed to invite users", err)
		} else if errors.Is(err, services.ErrAlreadyInRoom) {
			return utils.HandleError(c, fiber.StatusConflict, "User is already in the room", err)
		} else if errors.Is(err, services.ErrAlreadyInvited) {
			return utils.HandleError(c, fiber.StatusConflict, "User already has pending invite", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room / User not found")
	}

	return utils.HandleSuccess(c, "Invited users successfully", roomInvites)
}

// test if user not in room
func (h *RoomHandler) LeaveRoom(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	if err := h.roomService.LeaveRoom(roomId, userId); err != nil {
		if errors.Is(err, services.ErrRoomHasUnconsolidatedBills) {
			return utils.HandleError(c, fiber.StatusConflict, "Cannot leave room with unconsolidated bills", err)
		} else if errors.Is(err, services.ErrLeaveRoomAsHost) {
			return utils.HandleError(c, fiber.StatusConflict, "Host cannot leave room", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	return utils.HandleSuccess(c, "Left room successfully", nil)
}

func (h *RoomHandler) QueryVenue(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return utils.HandleError(c, fiber.StatusBadRequest, "Query parameter is required", nil)
	}

	venues, err := h.roomService.QueryVenue(query)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	return utils.HandleSuccess(c, "Queried venues successfully", venues)
}

// TODO: Implement endpoint for host to remove user from room
