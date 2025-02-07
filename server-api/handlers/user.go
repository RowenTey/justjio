package handlers

import (
	"fmt"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model/request"
	"github.com/RowenTey/JustJio/model/response"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"

	"github.com/gofiber/fiber/v2"
)

func GetUser(c *fiber.Ctx) error {
	id := c.Params("userId")
	user, err := (&services.UserService{DB: database.DB}).GetUserByID(id)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", id))
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
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", id))
	}

	return util.HandleSuccess(c, "User successfully updated", request)
}

func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("userId")

	userService := services.UserService{DB: database.DB}

	err := userService.DeleteUser(id)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", id))
	}
	return util.HandleSuccess(c, "User successfully deleted", nil)
}

func SendFriendRequest(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	var request request.ModifyFriendRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}
	if err := userService.SendFriendRequest(uint(userID), request.FriendID); err != nil {
		if err.Error() == "cannot send friend request to yourself" ||
			err.Error() == "already friends" ||
			err.Error() == "friend request already sent" {
			return util.HandleError(
				c, fiber.StatusConflict, err.Error(), err)
		}
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return util.HandleSuccess(c, "Friend request sent", nil)
}

func RemoveFriend(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	friendID, err := c.ParamsInt("friendId")
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}
	if err := userService.RemoveFriend(uint(userID), uint(friendID)); err != nil {
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return util.HandleSuccess(c, "Friend successfully removed", nil)
}

func GetFriends(c *fiber.Ctx) error {
	userID := c.Params("userId")

	userService := services.UserService{DB: database.DB}

	friends, err := userService.GetFriends(userID)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	return util.HandleSuccess(c, "Friends retrieved successfully", friends)
}

func IsFriend(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	var request request.ModifyFriendRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}

	isFriend := userService.IsFriend(uint(userID), request.FriendID)

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

func SearchFriends(c *fiber.Ctx) error {
	userID := c.Params("userId")
	query := c.Query("query")

	userService := services.UserService{DB: database.DB}

	friends, err := userService.SearchUsers(userID, query)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	return util.HandleSuccess(c, "Friends retrieved successfully", friends)
}

func GetFriendRequestsByStatus(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	status := c.Query("status")
	userService := services.UserService{DB: database.DB}

	requests, err := userService.GetFriendRequestsByStatus(uint(userID), status)
	if err != nil {
		if err.Error() == "invalid status" {
			return util.HandleInvalidInputError(c, err)
		}
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return util.HandleSuccess(c, "Friend requests retrieved successfully", requests)
}

func CountPendingFriendRequests(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}

	count, err := userService.CountPendingFriendRequests(uint(userID))
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	response := response.CountPendingRequestsResponse{
		Count: count,
	}
	return util.HandleSuccess(c, "Pending friend requests counted successfully", response)
}

func RespondToFriendRequest(c *fiber.Ctx) error {
	var request request.RespondToFriendRequestRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}

	switch request.Action {
	case "accept":
		if err := userService.AcceptFriendRequest(uint(request.RequestID)); err != nil {
			if err.Error() == "friend request already processed" {
				return util.HandleError(
					c, fiber.StatusConflict, err.Error(), err)
			}
			return util.HandleNotFoundOrInternalError(c, err, "Error processing friend request")
		}
		return util.HandleSuccess(c, "Friend request accepted successfully", nil)
	case "reject":
		if err := userService.RejectFriendRequest(uint(request.RequestID)); err != nil {
			if err.Error() == "friend request already processed" {
				return util.HandleError(
					c, fiber.StatusConflict, err.Error(), err)
			}
			return util.HandleNotFoundOrInternalError(c, err, "Error processing friend request")
		}
		return util.HandleSuccess(c, "Friend request rejected successfully", nil)
	default:
		return util.HandleInvalidInputError(c, fmt.Errorf("invalid action: must be 'accept' or 'reject'"))
	}
}
