package handlers

import (
	"errors"
	"fmt"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model/request"
	"github.com/RowenTey/JustJio/model/response"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetUser(c *fiber.Ctx) error {
	id := c.Params("userId")
	user, err := (&services.UserService{DB: database.DB}).GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(
				c, fiber.StatusNotFound, fmt.Sprintf("No user found with ID %s", id), err)
		}
		return util.HandleInternalServerError(c, err)
	}
	return util.HandleSuccess(c, "User found successfully", user)
}

func UpdateUser(c *fiber.Ctx) error {
	var request request.UpdateUserRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	id := c.Params("userId")
	userService := services.UserService{DB: database.DB}
	err := userService.UpdateUserField(id, request.Field, request.Value)
	if err != nil {
		if err.Error() == fmt.Sprintf("User field %s not supported for update", request.Field) {
			return util.HandleInvalidInputError(c, err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "User successfully updated", request)
}

func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("userId")

	userService := services.UserService{DB: database.DB}

	err := userService.DeleteUser(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(
				c, fiber.StatusNotFound, fmt.Sprintf("No user found with ID %s", id), err)
		}
		return util.HandleInternalServerError(c, err)
	}
	return util.HandleSuccess(c, "User successfully deleted", nil)
}

func AddFriend(c *fiber.Ctx) error {
	userID := c.Params("userId")
	var request request.ModifyFriendRequest

	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}

	if err := userService.AddFriend(userID, request.FriendID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(
				c, fiber.StatusNotFound, fmt.Sprintf("No user found with ID %s", userID), err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Friend successfully added", nil)
}

func RemoveFriend(c *fiber.Ctx) error {
	userID := c.Params("userId")
	var request request.ModifyFriendRequest

	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}

	if err := userService.RemoveFriend(userID, request.FriendID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(
				c, fiber.StatusNotFound, fmt.Sprintf("No user found with ID %s", userID), err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Friend successfully removed", nil)
}

func GetFriends(c *fiber.Ctx) error {
	userID := c.Params("userId")

	userService := services.UserService{DB: database.DB}

	friends, err := userService.GetFriends(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(
				c, fiber.StatusNotFound, fmt.Sprintf("No user found with ID %s", userID), err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Friends retrieved successfully", friends)
}

func IsFriend(c *fiber.Ctx) error {
	userID := c.Params("userId")
	var request struct {
		FriendID string `json:"frienId"`
	}

	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}

	isFriend, err := userService.IsFriend(userID, request.FriendID)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	response := response.IsFriendResponse{
		IsFriend: isFriend,
	}
	return util.HandleSuccess(c, "Friend check completed", response)
}

func GetNumFriends(c *fiber.Ctx) error {
	userID := c.Params("userId")

	userService := services.UserService{DB: database.DB}

	numFriends, err := userService.GetNumFriends(userID)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	response := response.GetNumFriendsResponse{
		NumFriends: numFriends,
	}
	return util.HandleSuccess(c, "Number of friends retrieved successfully", response)
}
