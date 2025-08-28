package handlers

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/dto/response"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService *services.UserService
	logger      *log.Entry
}

func NewUserHandler(userService *services.UserService, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      utils.AddServiceField(logger, "UserHandler"),
	}
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("userId")

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", id))
	}

	return utils.HandleSuccess(c, "User found successfully", user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	var request request.UpdateUserRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	id := c.Params("userId")
	if err := h.userService.UpdateUserField(id, request.Field, request.Value); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", id))
	}

	h.logger.Infof("User %s updated field %s to %s", id, request.Field, request.Value)
	return utils.HandleSuccess(c, "User successfully updated", request)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("userId")
	if err := h.userService.DeleteUser(id); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", id))
	}
	return utils.HandleSuccess(c, "User successfully deleted", nil)
}

func (h *UserHandler) SendFriendRequest(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	var request request.ModifyFriendRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if err := h.userService.SendFriendRequest(uint(userID), request.FriendID); err != nil {
		if errors.Is(err, services.ErrNoSelfFriendRequest) ||
			errors.Is(err, services.ErrAlreadyFriends) ||
			errors.Is(err, services.ErrFriendRequestExists) {
			return utils.HandleError(
				c, fiber.StatusConflict, err.Error(), err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return utils.HandleSuccess(c, "Friend request sent", nil)
}

func (h *UserHandler) RemoveFriend(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	friendID, err := c.ParamsInt("friendId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if err := h.userService.RemoveFriend(uint(userID), uint(friendID)); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return utils.HandleSuccess(c, "Friend successfully removed", nil)
}

func (h *UserHandler) GetFriends(c *fiber.Ctx) error {
	userID := c.Params("userId")

	friends, err := h.userService.GetFriends(userID)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	return utils.HandleSuccess(c, "Friends retrieved successfully", friends)
}

func (h *UserHandler) IsFriend(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	var request request.ModifyFriendRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	isFriend := h.userService.IsFriend(uint(userID), request.FriendID)

	response := response.IsFriendResponse{
		IsFriend: isFriend,
	}
	return utils.HandleSuccess(c, "Friend check completed", response)
}

func (h *UserHandler) GetNumFriends(c *fiber.Ctx) error {
	userID := c.Params("userId")

	numFriends, err := h.userService.GetNumFriends(userID)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	response := response.GetNumFriendsResponse{
		NumFriends: numFriends,
	}
	return utils.HandleSuccess(c, "Number of friends retrieved successfully", response)
}

func (h *UserHandler) SearchFriends(c *fiber.Ctx) error {
	userID := c.Params("userId")
	query := c.Query("query")

	friends, err := h.userService.SearchUsers(userID, query)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	return utils.HandleSuccess(c, "Friends retrieved successfully", friends)
}

func (h *UserHandler) GetFriendRequestsByStatus(c *fiber.Ctx) error {
	status := c.Query("status")
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	requests, err := h.userService.GetFriendRequestsByStatus(uint(userID), status)
	if err != nil {
		if errors.Is(err, services.ErrInvalidFriendRequestStatus) {
			return utils.HandleInvalidInputError(c, err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return utils.HandleSuccess(c, "Friend requests retrieved successfully", requests)
}

func (h *UserHandler) CountPendingFriendRequests(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	count, err := h.userService.CountPendingFriendRequests(uint(userID))
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	response := response.CountPendingRequestsResponse{
		Count: count,
	}
	return utils.HandleSuccess(c, "Pending friend requests counted successfully", response)
}

func (h *UserHandler) RespondToFriendRequest(c *fiber.Ctx) error {
	var request request.RespondToFriendRequestRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	requestIdUint := uint(request.RequestID)

	switch request.Action {
	case "accept":
		if err := h.userService.AcceptFriendRequest(requestIdUint); err != nil {
			if errors.Is(err, services.ErrFriendRequestAlreadyProcessed) {
				return utils.HandleError(
					c, fiber.StatusConflict, err.Error(), err)
			}
			return utils.HandleNotFoundOrInternalError(c, err, "Error processing friend request")
		}
		return utils.HandleSuccess(c, "Friend request accepted successfully", nil)
	case "reject":
		if err := h.userService.RejectFriendRequest(requestIdUint); err != nil {
			if errors.Is(err, services.ErrFriendRequestAlreadyProcessed) {
				return utils.HandleError(
					c, fiber.StatusConflict, err.Error(), err)
			}
			return utils.HandleNotFoundOrInternalError(c, err, "Error processing friend request")
		}
		return utils.HandleSuccess(c, "Friend request rejected successfully", nil)
	default:
		return utils.HandleInvalidInputError(c, fmt.Errorf("invalid action: must be 'accept' or 'reject'"))
	}
}
