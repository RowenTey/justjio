package handlers

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
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
// @Security BearerAuth
// @Router /rooms/{roomId} [get]
func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	roomId := c.Params("roomId")
	room, err := h.roomService.GetRoomById(ctx, roomId)
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
// @Success 200 {object} object{status=string,message=string,data=[]response.RoomListDto} "Retrieved rooms successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No rooms found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms [get]
func (h *RoomHandler) GetRooms(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")
	page := c.QueryInt("page", 1)

	rooms, err := h.roomService.GetRoomsByUserId(ctx, userId, page)
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
// @Success 200 {object} object{status=string,message=string,data=int} "Retrieved number of rooms successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No rooms found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/count [get]
func (h *RoomHandler) GetNumRooms(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	numRooms, err := h.roomService.GetNumRooms(ctx, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No rooms found")
	}

	return utils.HandleSuccess(c, "Retrieved number of rooms successfully", numRooms)
}

// GetUnjoinedPublicRooms retrieves public rooms the user hasn't joined
// @Summary Get unjoined public rooms
// @Description Retrieves public rooms that the authenticated user hasn't joined yet
// @Tags Rooms
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,data=[]response.SimplifiedRoomDto} "Retrieved public rooms successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No public rooms found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/public [get]
func (h *RoomHandler) GetUnjoinedPublicRooms(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	rooms, err := h.roomService.GetUnjoinedPublicRooms(ctx, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No public rooms found")
	}

	return utils.HandleSuccess(c, "Retrieved public rooms successfully", rooms)
}

// GetRoomInvites retrieves room invites for the user
// @Summary Get room invites
// @Description Retrieves pending room invites for the authenticated user
// @Tags Room Invites
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,data=[]response.RoomInviteDto} "Retrieved room invites successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No room invites found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/invites [get]
func (h *RoomHandler) GetRoomInvites(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	invites, err := h.roomService.GetRoomInvites(ctx, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No room invites found")
	}

	return utils.HandleSuccess(c, "Retrieved room invites successfully", invites)
}

// GetNumRoomInvites gets the count of room invites
// @Summary Get number of room invites
// @Description Retrieves the count of pending room invites for the authenticated user
// @Tags Room Invites
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,data=int} "Retrieved number of invites successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No room invites found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/invites/count [get]
func (h *RoomHandler) GetNumRoomInvites(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	numInvites, err := h.roomService.GetNumRoomInvites(ctx, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No room invites found")
	}

	return utils.HandleSuccess(c, "Retrieved number of invites successfully", numInvites)
}

// GetUninvitedFriendsForRoom retrieves friends not invited to a room
// @Summary Get uninvited friends for room
// @Description Retrieves the list of friends who haven't been invited to the specified room
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} object{status=string,message=string,data=[]response.MinimalUserDto} "Retrieved uninvited friends successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No uninvited friends found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/uninvited [get]
func (h *RoomHandler) GetUninvitedFriendsForRoom(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")
	roomId := c.Params("roomId")

	friends, err := h.roomService.GetUninvitedFriendsForRoom(ctx, roomId, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No uninvited friends found")
	}

	return utils.HandleSuccess(c, "Retrieved uninvited friends successfully", friends)
}

// CreateRoom creates a new room with invites
// @Summary Create room
// @Description Creates a new room and optionally sends invites to specified users
// @Tags Rooms
// @Accept json
// @Produce json
// @Param room body request.CreateRoomRequest true "Room creation details"
// @Success 200 {object} object{status=string,message=string,data=string} "Created room successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 500 {object} utils.EmptyApiResponse "Failed to create room and invites"
// @Security BearerAuth
// @Router /rooms [post]
func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	req := middleware.GetValidatedRequest[request.CreateRoomRequest](c)

	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	room := &model.Room{
		Name:         req.Name,
		Time:         req.Time,
		Venue:        req.Venue,
		VenuePlaceId: req.VenuePlaceId,
		Date:         req.Date,
		Description:  req.Description,
		IsPrivate:    req.IsPrivate,
		ImageUrl:     req.ImageUrl,
	}

	roomId, err := h.roomService.CreateRoomWithInvites(ctx, room, userId, req.Invitees)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Failed to create room and invites")
	}

	return utils.HandleSuccess(c, "Created room successfully", roomId)
}

// EditRoom updates room details
// @Summary Edit room
// @Description Updates details of an existing room (host only)
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param room body request.EditRoomRequest true "Room update details"
// @Success 200 {object} utils.EmptyApiResponse "Edited room successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 401 {object} utils.EmptyApiResponse "Only hosts can edit rooms"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/edit [patch]
func (h *RoomHandler) EditRoom(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	req := middleware.GetValidatedRequest[request.EditRoomRequest](c)

	roomId := c.Params("roomId")
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	if err := h.roomService.UpdateRoom(ctx, req, roomId, userId); err != nil {
		if errors.Is(err, services.ErrInvalidHost) {
			return utils.HandleError(c, fiber.StatusUnauthorized, "Only hosts can edit rooms", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	return utils.HandleSuccess[any](c, "Edited room successfully", nil)
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
	ctx := utils.GetOtelContext(c)
	roomId := c.Params("roomId")
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	if err := h.roomService.CloseRoom(ctx, roomId, userId); err != nil {
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

// JoinRoom joins a room
// @Summary Join room
// @Description Allows a user to join a room
// @Tags Rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} object{status=string,message=string,data=response.RoomDto} "Joined room successfully"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 409 {object} utils.EmptyApiResponse "User is already in room"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/join [patch]
func (h *RoomHandler) JoinRoom(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	room, err := h.roomService.JoinRoom(ctx, roomId, userId)
	if err != nil {
		if errors.Is(err, services.ErrAlreadyInRoom) {
			return utils.HandleError(c, fiber.StatusConflict, "User is already in room", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	h.logger.Info("User " + utils.GetUserInfoFromToken(token, "username") + " joined room " + roomId + " successfully.")
	return utils.HandleSuccess(c, "Joined room successfully", room)
}

// RespondToRoomInvite responds to a room invitation
// @Summary Respond to room invitation
// @Description Accepts or rejects a room invitation
// @Tags Room Invites
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param response body request.RespondToRoomInviteRequest true "Invitation response"
// @Success 200 {object} object{status=string,message=string,data=response.RoomDto} "Joined room successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId} [patch]
func (h *RoomHandler) RespondToRoomInvite(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	req := middleware.GetValidatedRequest[request.RespondToRoomInviteRequest](c)

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	room, err := h.roomService.RespondToRoomInvite(ctx, roomId, userId, req.Accept)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	if !req.Accept {
		return utils.HandleSuccess[any](c, "Rejected room invitation successfully", nil)
	}

	h.logger.Info(
		"User " + utils.GetUserInfoFromToken(token, "username") + " joined Room " + roomId + " successfully.")
	return utils.HandleSuccess(c, "Joined room successfully", room)
}

// InviteUser invites users to a room
// @Summary Invite users to room
// @Description Invites multiple users to a room (host only)
// @Tags Room Invites
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param invites body request.InviteUserRequest true "User invites details"
// @Success 200 {object} utils.EmptyApiResponse "Invited users successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 403 {object} utils.EmptyApiResponse "Only hosts are allowed to invite users"
// @Failure 404 {object} utils.EmptyApiResponse "Room / User not found"
// @Failure 409 {object} utils.EmptyApiResponse "User is already in the room or already has pending invite"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId} [post]
func (h *RoomHandler) InviteUser(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	req := middleware.GetValidatedRequest[request.InviteUserRequest](c)

	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")
	roomId := c.Params("roomId")

	err := h.roomService.InviteUsersToRoom(ctx, roomId, userId, req.Invitees)
	if err != nil {
		if errors.Is(err, services.ErrInvalidHost) {
			return utils.HandleError(c, fiber.StatusForbidden, "Only hosts are allowed to invite users", err)
		} else if errors.Is(err, services.ErrAlreadyInRoom) {
			return utils.HandleError(c, fiber.StatusConflict, "User is already in the room", err)
		} else if errors.Is(err, services.ErrAlreadyInvited) {
			return utils.HandleError(c, fiber.StatusConflict, "User already has pending invite", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room / User not found")
	}

	return utils.HandleSuccess[any](c, "Invited users successfully", nil)
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
	ctx := utils.GetOtelContext(c)
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")
	roomId := c.Params("roomId")

	if err := h.roomService.LeaveRoom(ctx, roomId, userId); err != nil {
		if errors.Is(err, services.ErrRoomHasUnconsolidatedBills) {
			return utils.HandleError(c, fiber.StatusConflict, "Cannot leave room with unconsolidated bills", err)
		} else if errors.Is(err, services.ErrLeaveRoomAsHost) {
			return utils.HandleError(c, fiber.StatusConflict, "Host cannot leave room", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	return utils.HandleSuccess[any](c, "Left room successfully", nil)
}

// TODO: Move to location handler
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
// @Security BearerAuth
// @Router /roooms/venues/search [get]
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
