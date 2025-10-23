package handlers

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"

	"github.com/gofiber/fiber/v2"
)

var (
	validStatuses = map[string]bool{"pending": true, "accepted": true, "rejected": true}
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

// GetUser retrieves a user by ID
// @Summary Get user details
// @Description Retrieves a user's details based on the provided user ID
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} object{status=string,message=string,data=model.User} "User found successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	id := c.Params("userId")

	user, err := h.userService.GetUserByID(ctx, id)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", id))
	}

	return utils.HandleSuccess(c, "User found successfully", user)
}

// UpdateUsername updates a user's username
// @Summary Update username
// @Description Updates the username for a specific user
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "User ID" example("123")
// @Param request body request.UpdateUsernameRequest true "Username update request" example({"username":"new_username"})
// @Success 200 {object} utils.EmptyApiResponse "User successfully updated"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 409 {object} utils.EmptyApiResponse "Username already taken"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/username [patch]
func (h *UserHandler) UpdateUsername(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	req := middleware.GetValidatedRequest[request.UpdateUsernameRequest](c)

	id := c.Params("userId")
	if err := h.userService.UpdateUsername(ctx, id, req.Username); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return utils.HandleError(
				c, fiber.StatusConflict, fmt.Sprintf("Username '%s' is already taken", req.Username), nil)
		}
		return utils.HandleNotFoundOrInternalError(c, err,
			fmt.Sprintf("No user found with ID %s", id))
	}

	h.logger.Infof("User %s updated username to %s", id, req.Username)
	return utils.HandleSuccess[any](c, "User successfully updated", nil)
}

// GetNumFriends gets the number of friends for a user
// @Summary Get number of friends
// @Description Retrieves the total count of friends for a specific user
// @Tags Friends
// @Accept json
// @Produce json
// @Param userId path string true "User ID" example("123")
// @Success 200 {object} object{status=string,message=string,data=int} "Number of friends retrieved successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/friends/count [get]
func (h *UserHandler) GetNumFriends(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userID := c.Params("userId")

	numFriends, err := h.userService.GetNumFriends(ctx, userID)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	return utils.HandleSuccess(c, "Number of friends retrieved successfully", numFriends)
}

// RemoveFriend removes a friend
// @Summary Remove friend
// @Description Removes a friend relationship between two users
// @Tags Friends
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Param friendId path int true "Friend ID to remove"
// @Success 200 {object} utils.EmptyApiResponse "Friend successfully removed"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/friends/{friendId} [delete]
func (h *UserHandler) RemoveFriend(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	friendID, err := c.ParamsInt("friendId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if err := h.userService.RemoveFriend(ctx, uint(userID), uint(friendID)); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return utils.HandleSuccess[any](c, "Friend successfully removed", nil)
}

// GetFriends retrieves a user's friends
// @Summary Get friends
// @Description Retrieves the list of friends for a user
// @Tags Friends
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} object{status=string,message=string,data=[]response.MinimalUserDto} "Friends retrieved successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/friends [get]
func (h *UserHandler) GetFriends(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userID := c.Params("userId")

	friends, err := h.userService.GetFriends(ctx, userID)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	return utils.HandleSuccess(c, "Friends retrieved successfully", friends)
}

// SearchNonFriends searches for non-friend users
// @Summary Search non-friend users
// @Description Searches for users who are not friends with the specified user based on a query
// @Tags Friends
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param query query string true "Search query"
// @Success 200 {object} object{status=string,message=string,data=[]response.MinimalUserDto} "Non-friend users retrieved successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/friends/search [get]
func (h *UserHandler) SearchNonFriends(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userID := c.Params("userId")
	query := c.Query("query")

	friends, err := h.userService.SearchNonFriendUsers(ctx, userID, query)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %s", userID))
	}

	return utils.HandleSuccess(c, "Non-friend users retrieved successfully", friends)
}

// SendFriendRequest sends a friend request
// @Summary Send friend request
// @Description Sends a friend request from one user to another
// @Tags Friends
// @Accept json
// @Produce json
// @Param userId path int true "User ID of the sender"
// @Param request body request.ModifyFriendRequest true "Friend request details"
// @Success 200 {object} utils.EmptyApiResponse "Friend request sent"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 409 {object} utils.EmptyApiResponse "Conflict: Self request, already friends, or request exists"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/friendRequests [post]
func (h *UserHandler) SendFriendRequest(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	req := middleware.GetValidatedRequest[request.ModifyFriendRequest](c)

	if err := h.userService.SendFriendRequest(ctx, uint(userID), req.FriendID); err != nil {
		if errors.Is(err, services.ErrNoSelfFriendRequest) ||
			errors.Is(err, services.ErrAlreadyFriends) ||
			errors.Is(err, services.ErrFriendRequestExists) {
			return utils.HandleError(
				c, fiber.StatusConflict, err.Error(), err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return utils.HandleSuccess[any](c, "Friend request sent", nil)
}

// GetFriendRequestsByStatus retrieves friend requests by status
// @Summary Get friend requests by status
// @Description Retrieves friend requests for a user filtered by status (e.g., pending, accepted)
// @Tags Friend Requests
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Param status query string true "Status of friend requests (e.g., pending, accepted)"
// @Success 200 {object} object{status=string,message=string,data=[]response.FriendRequestDto} "Friend requests retrieved successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid status"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/friendRequests [get]
func (h *UserHandler) GetFriendRequestsByStatus(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)

	status := c.Query("status")
	if !validStatuses[status] {
		return utils.HandleInvalidInputError(c, errors.New("invalid status"))
	}

	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	requests, err := h.userService.GetFriendRequestsByStatus(ctx, uint(userID), status)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return utils.HandleSuccess(c, "Friend requests retrieved successfully", requests)
}

// CountPendingFriendRequests counts pending friend requests
// @Summary Count pending friend requests
// @Description Counts the number of pending friend requests for a user
// @Tags Friend Requests
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} object{status=string,message=string,data=int} "Pending friend requests counted successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 404 {object} utils.EmptyApiResponse "No user found with ID"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/friendRequests/count [get]
func (h *UserHandler) CountPendingFriendRequests(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)

	userID, err := c.ParamsInt("userId")
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	count, err := h.userService.CountPendingFriendRequests(ctx, uint(userID))
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, fmt.Sprintf("No user found with ID %d", userID))
	}

	return utils.HandleSuccess(c, "Pending friend requests counted successfully", count)
}

// RespondToFriendRequest responds to a friend request
// @Summary Respond to friend request
// @Description Accepts or rejects a friend request based on the provided action
// @Tags Friend Requests
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Param request body request.RespondToFriendRequestRequest true "Friend request response details"
// @Success 200 {object} utils.EmptyApiResponse "Friend request processed successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input or action"
// @Failure 404 {object} utils.EmptyApiResponse "Error processing friend request"
// @Failure 409 {object} utils.EmptyApiResponse "Friend request already processed"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /users/{userId}/friendRequests [patch]
func (h *UserHandler) RespondToFriendRequest(c *fiber.Ctx) error {
	ctx := utils.GetOtelContext(c)
	req := middleware.GetValidatedRequest[request.RespondToFriendRequestRequest](c)

	requestIdUint := uint(req.RequestID)

	switch req.Action {
	case "accept":
		if err := h.userService.AcceptFriendRequest(ctx, requestIdUint); err != nil {
			if errors.Is(err, services.ErrFriendRequestAlreadyProcessed) {
				return utils.HandleError(
					c, fiber.StatusConflict, err.Error(), err)
			}
			return utils.HandleNotFoundOrInternalError(c, err, "Error processing friend request")
		}
		return utils.HandleSuccess[any](c, "Friend request accepted successfully", nil)
	case "reject":
		if err := h.userService.RejectFriendRequest(ctx, requestIdUint); err != nil {
			if errors.Is(err, services.ErrFriendRequestAlreadyProcessed) {
				return utils.HandleError(
					c, fiber.StatusConflict, err.Error(), err)
			}
			return utils.HandleNotFoundOrInternalError(c, err, "Error processing friend request")
		}
		return utils.HandleSuccess[any](c, "Friend request rejected successfully", nil)
	default:
		return utils.HandleInvalidInputError(c, fmt.Errorf("invalid action: must be 'accept' or 'reject'"))
	}
}
