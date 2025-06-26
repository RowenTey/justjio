package handlers

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/model/response"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService *services.AuthService
	// Store OTP with email as key
	clientOtpMap sync.Map
	logger       *log.Entry
}

func NewAuthHandler(
	authService *services.AuthService,
	logger *log.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		clientOtpMap: sync.Map{},
		logger:       utils.AddServiceField(logger, "AuthHandler"),
	}
}

func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}
	h.logger.Info("Received sign up request for user: ", user.Username)

	createdUser, err := h.authService.SignUp(&user, &h.clientOtpMap)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return utils.HandleError(
				c, fiber.StatusConflict, "Username or email already exists", err)
		}
		return utils.HandleInternalServerError(c, err)
	}

	response := response.AuthResponse{
		Email:      createdUser.Email,
		Username:   createdUser.Username,
		PictureUrl: createdUser.PictureUrl,
		UID:        createdUser.ID,
	}
	h.logger.Info("User " + response.Username + " signed up successfully.")
	return utils.HandleSuccess(c, "User signed up successfully", response)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input request.LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token, user, err := h.authService.Login(input.Username, input.Password)
	if err != nil {
		if errors.Is(err, services.ErrPasswordDoesNotMatch) {
			return utils.HandleError(c, fiber.StatusUnauthorized, "Invalid username or password", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	response := response.AuthResponse{
		Username:   user.Username,
		Email:      user.Email,
		PictureUrl: user.PictureUrl,
		UID:        user.ID,
	}
	h.logger.Info("User " + response.Username + " logged in successfully.")
	return utils.HandleLoginSuccess(c, "Login successfully", token, response)
}

func (h *AuthHandler) SendOTPEmail(c *fiber.Ctx) error {
	var request request.SendOTPEmailRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if err := h.
		authService.
		GenerateAndSendOTPEmail(request.Email, request.Purpose, &h.clientOtpMap); err != nil {
		if errors.Is(err, services.ErrInvalidPurpose) {
			return utils.HandleError(c, fiber.StatusBadRequest, "Invalid purpose", err)
		} else if errors.Is(err, services.ErrEmailAlreadyVerified) {
			return utils.HandleError(c, fiber.StatusBadRequest, "Email already verified", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	h.logger.Info("OTP sent to " + request.Email + " successfully.")
	return utils.HandleSuccess(c, "OTP sent successfully", nil)
}

func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var request request.VerifyOTPRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	err := h.authService.VerifyOTP(request.Email, request.OTP, &h.clientOtpMap)
	if err != nil {
		if errors.Is(err, services.ErrInvalidOTP) {
			return utils.HandleError(c, fiber.StatusBadRequest, "Invalid OTP", err)
		} else if errors.Is(err, services.ErrOTPNotFound) {
			return utils.HandleError(c, fiber.StatusNotFound, "OTP not found", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	h.logger.Println("OTP verified successfully for email", request.Email)
	return utils.HandleSuccess(c, "OTP verified successfully", nil)
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var request request.ResetPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if err := h.
		authService.
		ResetPassword(request.Email, request.Password); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	return utils.HandleSuccess(c, "Password reset successfully", nil)
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	var request request.GoogleAuthRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token, user, err := h.authService.GoogleLogin(request.Code)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	response := response.AuthResponse{
		Email:      user.Email,
		Username:   user.Username,
		PictureUrl: user.PictureUrl,
		UID:        user.ID,
	}

	h.logger.Println("User " + response.Username + " authenticated via Google OAuth")
	return utils.HandleLoginSuccess(c, "Authenticated via Google successfully", token, response)
}
