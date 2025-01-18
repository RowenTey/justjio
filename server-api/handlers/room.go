package handlers

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model/request"
	"github.com/RowenTey/JustJio/model/response"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func GetRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	room, err := (&services.RoomService{DB: database.DB}).GetRoomById(roomId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "Room not found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Retrieved room successfully", room)
}

func GetRooms(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	page := c.QueryInt("page", 1)

	roomService := &services.RoomService{DB: database.DB}

	rooms, err := roomService.GetRooms(userId, page)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "No rooms found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Retrieved rooms successfully", rooms)
}

func GetRoomInvitations(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")

	roomService := &services.RoomService{DB: database.DB}

	invites, err := roomService.GetRoomInvites(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "No room invitations found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Retrieved room invitations successfully", invites)
}

func GetRoomAttendees(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	roomService := &services.RoomService{DB: database.DB}

	attendees, err := roomService.GetRoomAttendees(roomId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "No attendees found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Retrieved room attendees successfully", attendees)
}

func CreateRoom(c *fiber.Ctx) error {
	var request request.CreateRoomRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")

	var inviteesIds []string
	json.Unmarshal([]byte(request.InviteesId), &inviteesIds)

	tx := database.DB.Begin()

	userService := &services.UserService{DB: tx}
	roomService := &services.RoomService{DB: tx}

	user, err := userService.GetUserByID(userId)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "User not found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	invitees, err := userService.ValidateUsers(inviteesIds)
	if err != nil {
		tx.Rollback()
		return util.HandleError(c, fiber.StatusNotFound, "User doesn't exist", err)
	}

	room, err := roomService.CreateRoom(&request.Room, user)
	if err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	invites, err := roomService.InviteUserToRoom(
		room.ID, user, invitees, request.Message)
	if err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	response := response.CreateRoomResponse{
		Room:    *room,
		Invites: *invites,
	}

	log.Println("Room " + room.Name + " created successfully.")
	return util.HandleSuccess(c, "Created room successfully", response)
}

func CloseRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")

	err := (&services.RoomService{DB: database.DB}).CloseRoom(roomId, userId)
	if err != nil {
		if err.Error() == "User is not the host of the room" {
			return util.HandleError(
				c, fiber.StatusUnauthorized, "Only hosts are allowed to close rooms", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Closed room successfully", nil)
}

func RespondToRoomInvite(c *fiber.Ctx) error {
	var request request.RespondToRoomInviteRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	roomService := &services.RoomService{DB: database.DB}

	status := "accepted"
	if !request.Accept {
		status = "rejected"
	}

	err := roomService.UpdateRoomInviteStatus(roomId, userId, status)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "Room not found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	if status == "rejected" {
		return util.HandleSuccess(c, "Rejected room invitation successfully", nil)
	}

	room, err := roomService.GetRoomById(roomId)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	attendees, err := roomService.GetRoomAttendees(roomId)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	roomResponse := response.JoinRoomResponse{
		Room:     *room,
		Attendes: *attendees,
	}

	log.Println(
		"User " + util.GetUserInfoFromToken(token, "username") + " joined Room " + roomId + " successfully.")
	return util.HandleSuccess(c, "Joined room successfully", roomResponse)
}

func InviteUser(c *fiber.Ctx) error {
	var request request.InviteUserRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	// TODO: Check if user is host of room

	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	var inviteesIds []string
	json.Unmarshal([]byte(request.InviteesId), &inviteesIds)

	tx := database.DB.Begin()
	userService := &services.UserService{DB: tx}

	user, err := userService.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "User not found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	invitees, err := userService.ValidateUsers(inviteesIds)
	if err != nil {
		return util.HandleError(c, fiber.StatusNotFound, "User doesn't exist", err)
	}

	roomInvites, err := (&services.RoomService{DB: tx}).InviteUserToRoom(
		roomId, user, invitees, "You have been invited to join this room")
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "Room / User not found", err)
		} else if err.Error() == "User is not the host of the room" {
			return util.HandleError(c, fiber.StatusUnauthorized, "Only hosts are allowed to invite users", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Invited users successfully", roomInvites)
}

func LeaveRoom(c *fiber.Ctx) error {
	var err error
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	tx := database.DB.Begin()
	err = (&services.RoomService{DB: tx}).RemoveUserFromRoom(roomId, userId)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "Room not found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Left room successfully", nil)
}

// TODO: Implement endpoint for host to remove user from room
