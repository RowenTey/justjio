package handlers

import (
	"errors"
	"fmt"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model/request"
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

// UpdateUser godoc
// @Summary      Update user attribute
// @Description  Update user attribute with new value
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        field   body      string  true  "Field"
// @Param        value   body      string  true  "Value"
// @Success      200  {object}  handlers.UpdateUser.UpdateUserInput
// @Failure      400  {object}  nil
// @Failure      401  {object}  nil
// @Failure      404  {object}  nil
// @Router       /users/{id} [patch]
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

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Delete a user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        password   body      handlers.DeleteUser.PasswordInput  true  "User Password"
// @Success      200  {object}  nil
// @Failure      400  {object}  nil
// @Failure      401  {object}  nil
// @Router       /users/{id} [delete]
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

// AddFriend godoc
// @Summary      Add a friend
// @Description  Add a friend to the user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id         path      int  true  "User ID"
// @Param        friend_id   body      string  true  "Friend ID"
// @Success      200  {object}  nil
// @Failure      400  {object}  nil
// @Failure      404  {object}  nil
// @Router       /users/{id}/friends [post]
func AddFriend(c *fiber.Ctx) error {
	userID := c.Params("userId")
	var request struct {
		FriendID string `json:"friend_id"`
	}

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

// RemoveFriend godoc
// @Summary      Remove a friend
// @Description  Remove a friend from the user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id         path      int  true  "User ID"
// @Param        friend_id   body      string  true  "Friend ID"
// @Success      200  {object}  nil
// @Failure      400  {object}  nil
// @Failure      404  {object}  nil
// @Router       /users/{id}/friends [delete]
func RemoveFriend(c *fiber.Ctx) error {
	userID := c.Params("userId")
	var request struct {
		FriendID string `json:"friend_id"`
	}

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

// GetFriends godoc
// @Summary      Get user friends
// @Description  Get all friends of a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  []model.User
// @Failure      404  {object}  nil
// @Router       /users/{id}/friends [get]
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

// IsFriend godoc
// @Summary      Check if users are friends
// @Description  Check if two users are friends
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id         path      int  true  "User ID"
// @Param        friend_id   body      string  true  "Friend ID"
// @Success      200  {object}  map[string]bool
// @Failure      400  {object}  nil
// @Failure      404  {object}  nil
// @Router       /users/{id}/friends/check [post]
func IsFriend(c *fiber.Ctx) error {
	userID := c.Params("userId")
	var request struct {
		FriendID string `json:"friend_id"`
	}

	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userService := services.UserService{DB: database.DB}

	isFriend, err := userService.IsFriend(userID, request.FriendID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(
				c, fiber.StatusNotFound, fmt.Sprintf("No user found with ID %s", userID), err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Friend check completed", map[string]bool{"isFriend": isFriend})
}
