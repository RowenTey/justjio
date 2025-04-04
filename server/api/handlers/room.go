package handlers

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/model/response"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var roomLogger = log.WithFields(log.Fields{"service": "RoomHandler"})

func GetRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	room, err := services.NewRoomService(database.DB).GetRoomById(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	return utils.HandleSuccess(c, "Retrieved room successfully", room)
}

func GetRooms(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	page := c.QueryInt("page", 1)

	roomService := services.NewRoomService(database.DB)

	rooms, err := roomService.GetRooms(userId, page)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No rooms found")
	}

	return utils.HandleSuccess(c, "Retrieved rooms successfully", rooms)
}

func GetNumRooms(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	roomService := services.NewRoomService(database.DB)

	numRooms, err := roomService.GetNumRooms(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No rooms found")
	}

	response := response.GetNumRoomsResponse{Count: int(numRooms)}
	return utils.HandleSuccess(c, "Retrieved number of rooms successfully", response)
}

func GetRoomInvitations(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	roomService := services.NewRoomService(database.DB)

	invites, err := roomService.GetRoomInvites(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No room invitations found")
	}

	return utils.HandleSuccess(c, "Retrieved room invitations successfully", invites)
}

func GetNumRoomInvitations(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	roomService := services.NewRoomService(database.DB)

	numInvites, err := roomService.GetNumRoomInvites(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No room invitations found")
	}

	response := response.GetNumRoomInvitationsResponse{Count: int(numInvites)}
	return utils.HandleSuccess(c, "Retrieved number of invitations successfully", response)
}

func GetRoomAttendees(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	roomService := services.NewRoomService(database.DB)

	attendees, err := roomService.GetRoomAttendees(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No attendees found")
	}

	return utils.HandleSuccess(c, "Retrieved room attendees successfully", attendees)
}

func GetUninvitedFriendsForRoom(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	roomService := services.NewRoomService(database.DB)
	friends, err := roomService.GetUninvitedFriendsForRoom(roomId, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No uninvited friends found")
	}

	return utils.HandleSuccess(c, "Retrieved uninvited friends successfully", friends)
}

func CreateRoom(c *fiber.Ctx) error {
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

	tx := database.DB.Begin()

	userService := services.NewUserService(tx)
	roomService := services.NewRoomService(tx)

	user, err := userService.GetUserByID(userId)
	if err != nil {
		tx.Rollback()
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	invitees, err := userService.ValidateUsers(inviteesIds)
	if err != nil {
		tx.Rollback()
		return utils.HandleError(c, fiber.StatusNotFound, "User doesn't exist", err)
	}

	room, err := roomService.CreateRoom(&request.Room, user)
	if err != nil {
		tx.Rollback()
		return utils.HandleInternalServerError(c, err)
	}

	invites, err := roomService.InviteUserToRoom(
		room.ID, user, invitees, request.Message)
	if err != nil {
		tx.Rollback()
		return utils.HandleInternalServerError(c, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.HandleInternalServerError(c, err)
	}

	response := response.CreateRoomResponse{
		Room:    *room,
		Invites: *invites,
	}

	roomLogger.Info("Room " + room.Name + " created successfully.")
	return utils.HandleSuccess(c, "Created room successfully", response)
}

func CloseRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	// TODO: user can close room if they are not any bills to consolidate
	// check if they are unconsolidated bills
	consolidated, err := services.NewBillService(database.DB).IsRoomBillConsolidated(roomId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.HandleInternalServerError(c, err)
	}

	if !consolidated {
		return utils.HandleError(
			c, fiber.StatusConflict, "Cannot close room with unconsolidated bills", nil)
	}

	err = services.NewRoomService(database.DB).CloseRoom(roomId, userId)
	if err != nil {
		if err.Error() == "user is not the host of the room" {
			return utils.HandleError(
				c, fiber.StatusUnauthorized, "Only hosts are allowed to close rooms", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	return utils.HandleSuccess(c, "Closed room successfully", nil)
}

func JoinRoom(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	roomService := services.NewRoomService(database.DB)

	err := roomService.JoinRoom(roomId, userId)
	if err != nil {
		if err.Error() == "user is already in room" {
			return utils.HandleError(c, fiber.StatusConflict, "User is already in room", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	room, err := roomService.GetRoomById(roomId)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	attendees, err := roomService.GetRoomAttendees(roomId)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	roomResponse := response.JoinRoomResponse{
		Room:      *room,
		Attendees: *attendees,
	}

	roomLogger.Info("User " + utils.GetUserInfoFromToken(token, "username") + " joined Room " + roomId + " successfully.")
	return utils.HandleSuccess(c, "Joined room successfully", roomResponse)
}

func RespondToRoomInvite(c *fiber.Ctx) error {
	var request request.RespondToRoomInviteRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	roomService := services.NewRoomService(database.DB)

	status := "accepted"
	if !request.Accept {
		status = "rejected"
	}

	err := roomService.UpdateRoomInviteStatus(roomId, userId, status)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	if status == "rejected" {
		return utils.HandleSuccess(c, "Rejected room invitation successfully", nil)
	}

	room, err := roomService.GetRoomById(roomId)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	attendees, err := roomService.GetRoomAttendees(roomId)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	roomResponse := response.JoinRoomResponse{
		Room:      *room,
		Attendees: *attendees,
	}

	roomLogger.Info(
		"User " + utils.GetUserInfoFromToken(token, "username") + " joined Room " + roomId + " successfully.")
	return utils.HandleSuccess(c, "Joined room successfully", roomResponse)
}

func InviteUser(c *fiber.Ctx) error {
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

	tx := database.DB.Begin()

	userService := services.NewUserService(tx)

	user, err := userService.GetUserByID(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	invitees, err := userService.ValidateUsers(inviteesIds)
	if err != nil {
		return utils.HandleError(c, fiber.StatusNotFound, "User doesn't exist", err)
	}

	roomInvites, err := services.NewRoomService(tx).InviteUserToRoom(
		roomId, user, invitees, request.Message)
	if err != nil {
		tx.Rollback()
		if err.Error() == "user is not the host of the room" {
			return utils.HandleError(c, fiber.StatusUnauthorized, "Only hosts are allowed to invite users", err)
		} else if err.Error() == "user is already in the room" {
			return utils.HandleError(c, fiber.StatusConflict, "User is already in the room", err)
		} else if err.Error() == "user already has pending invite" {
			return utils.HandleError(c, fiber.StatusConflict, "User already has pending invite", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Room / User not found")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.HandleInternalServerError(c, err)
	}

	return utils.HandleSuccess(c, "Invited users successfully", roomInvites)
}

func LeaveRoom(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	consolidated, err := services.NewBillService(database.DB).IsRoomBillConsolidated(roomId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.HandleInternalServerError(c, err)
	}

	if !consolidated {
		return utils.HandleError(
			c, fiber.StatusConflict, "Cannot leave room with unconsolidated bills", nil)
	}

	// TODO: check that user is not the host of the room
	if err := services.NewRoomService(database.DB).RemoveUserFromRoom(roomId, userId); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	return utils.HandleSuccess(c, "Left room successfully", nil)
}

// TODO: Implement endpoint for host to remove user from room
