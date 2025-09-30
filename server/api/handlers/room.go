package handlers

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/dto/response"
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

// GetRoom retrieves a room by ID
// @Summary Get room details
// @Description Retrieves details of a specific room by its ID
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} object{status=string,message=string,data=response.RoomDto} "Retrieved room successfully"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /rooms/{roomId} [get]
func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	room, err := h.roomService.GetRoomById(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}
	return utils.HandleSuccess(c, "Retrieved room successfully", room)
}

// GetRooms retrieves rooms for the authenticated user
// @Summary Get user's rooms
// @Description Retrieves a paginated list of rooms for the authenticated user
// @Tags Rooms
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object{status=string,message=string,data=[]model.Room} "Retrieved rooms successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No rooms found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms [get]
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

// GetNumRooms gets the count of user's rooms
// @Summary Get number of rooms
// @Description Retrieves the total count of rooms for the authenticated user
// @Tags Rooms
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,data=response.GetNumRoomsResponse} "Retrieved number of rooms successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No rooms found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/count [get]
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

// GetUnjoinedPublicRooms retrieves public rooms the user hasn't joined
// @Summary Get unjoined public rooms
// @Description Retrieves public rooms that the authenticated user hasn't joined yet
// @Tags Rooms
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,data=[]model.Room} "Retrieved public rooms successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No public rooms found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/public [get]
func (h *RoomHandler) GetUnjoinedPublicRooms(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	rooms, err := h.roomService.GetUnjoinedPublicRooms(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No public rooms found")
	}

	return utils.HandleSuccess(c, "Retrieved public rooms successfully", rooms)
}

// GetRoomInvitations retrieves room invitations for the user
// @Summary Get room invitations
// @Description Retrieves pending room invitations for the authenticated user
// @Tags Room Invitations
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,data=[]model.RoomInvite} "Retrieved room invitations successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No room invitations found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/invitations [get]
func (h *RoomHandler) GetRoomInvitations(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	invites, err := h.roomService.GetRoomInvites(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No room invitations found")
	}

	return utils.HandleSuccess(c, "Retrieved room invitations successfully", invites)
}

// GetNumRoomInvitations gets the count of room invitations
// @Summary Get number of room invitations
// @Description Retrieves the count of pending room invitations for the authenticated user
// @Tags Room Invitations
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,data=response.GetNumRoomInvitationsResponse} "Retrieved number of invitations successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No room invitations found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/invitations/count [get]
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

// GetRoomAttendees retrieves attendees of a room
// @Summary Get room attendees
// @Description Retrieves the list of attendees for a specific room
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} object{status=string,message=string,data=[]model.User} "Retrieved room attendees successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No attendees found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /rooms/{roomId}/attendees [get]
func (h *RoomHandler) GetRoomAttendees(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	attendees, err := h.roomService.GetRoomAttendees(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No attendees found")
	}

	return utils.HandleSuccess(c, "Retrieved room attendees successfully", attendees)
}

// GetUninvitedFriendsForRoom retrieves friends not invited to a room
// @Summary Get uninvited friends for room
// @Description Retrieves the list of friends who haven't been invited to the specified room
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} object{status=string,message=string,data=[]model.User} "Retrieved uninvited friends successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No uninvited friends found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/uninvited-friends [get]
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

// CreateRoom creates a new room with invitations
// @Summary Create room
// @Description Creates a new room and optionally sends invitations to specified users
// @Tags Rooms
// @Accept json
// @Produce json
// @Param room body request.CreateRoomRequest true "Room creation details"
// @Success 200 {object} object{status=string,message=string,data=response.CreateRoomResponse} "Created room successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 500 {object} utils.EmptyApiResponse "Failed to create room and invites"
// @Security BearerAuth
// @Router /rooms [post]
func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	var request request.CreateRoomRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

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

// EditRoom updates room details
// @Summary Edit room
// @Description Updates details of an existing room (host only)
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param room body request.UpdateRoomRequest true "Room update details"
// @Success 200 {object} object{status=string,message=string,data=model.Room} "Edited room successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 401 {object} utils.EmptyApiResponse "Only hosts can edit rooms"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId} [patch]
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
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	h.logger.Info("Room " + room.Name + " edited successfully.")
	return utils.HandleSuccess(c, "Edited room successfully", room)
}

// CloseRoom closes a room
// @Summary Close room
// @Description Closes a room (host only, no unconsolidated bills allowed)
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} utils.EmptyApiResponse "Closed room successfully"
// @Failure 401 {object} utils.EmptyApiResponse "Only hosts are allowed to close rooms"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 409 {object} utils.EmptyApiResponse "Cannot close room with unconsolidated bills"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/close [patch]
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

	return utils.HandleSuccess[any](c, "Closed room successfully", nil)
}

// JoinRoom joins a public room
// @Summary Join room
// @Description Allows a user to join a public room
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} object{status=string,message=string,data=response.JoinRoomResponse} "Joined room successfully"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 409 {object} utils.EmptyApiResponse "User is already in room"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/join [post]
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

// RespondToRoomInvite responds to a room invitation
// @Summary Respond to room invitation
// @Description Accepts or rejects a room invitation
// @Tags Room Invitations
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param response body request.RespondToRoomInviteRequest true "Invitation response"
// @Success 200 {object} object{status=string,message=string,data=response.JoinRoomResponse} "Joined room successfully"
// @Success 200 {object} utils.EmptyApiResponse "Rejected room invitation successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/invitations/respond [patch]
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
		return utils.HandleSuccess[any](c, "Rejected room invitation successfully", nil)
	}

	roomResponse := response.JoinRoomResponse{
		Room:      *room,
		Attendees: *attendees,
	}
	h.logger.Info(
		"User " + utils.GetUserInfoFromToken(token, "username") + " joined Room " + roomId + " successfully.")
	return utils.HandleSuccess(c, "Joined room successfully", roomResponse)
}

// InviteUser invites users to a room
// @Summary Invite users to room
// @Description Invites multiple users to a room (host only)
// @Tags Room Invitations
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param invites body request.InviteUserRequest true "User invitation details"
// @Success 200 {object} object{status=string,message=string,data=[]model.RoomInvite} "Invited users successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 401 {object} utils.EmptyApiResponse "Only hosts are allowed to invite users"
// @Failure 404 {object} utils.EmptyApiResponse "Room / User not found"
// @Failure 409 {object} utils.EmptyApiResponse "User is already in the room or already has pending invite"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/invitations [post]
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

// LeaveRoom leaves a room
// @Summary Leave room
// @Description Allows a user to leave a room (not allowed for hosts or with unconsolidated bills)
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} utils.EmptyApiResponse "Left room successfully"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 409 {object} utils.EmptyApiResponse "Cannot leave room with unconsolidated bills or host cannot leave room"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/leave [delete]
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

	return utils.HandleSuccess[any](c, "Left room successfully", nil)
}

// QueryVenue searches for venues
// @Summary Query venues
// @Description Searches for venues based on a query string
// @Tags Venues
// @Accept json
// @Produce json
// @Param query query string true "Search query for venues"
// @Success 200 {object} object{status=string,message=string,data=[]model_location.Venue} "Queried venues successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Query parameter is required"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /venues/search [get]
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
