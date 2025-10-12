package handlers

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/dto/response"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService *services.AuthService
	// Store OTP with email as key
	ClientOtpMap sync.Map
	logger       *log.Entry
}

func NewAuthHandler(
	authService *services.AuthService,
	logger *log.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		ClientOtpMap: sync.Map{},
		logger:       utils.AddServiceField(logger, "AuthHandler"),
	}
}

// SignUp creates a new user account
// @Summary Sign up
// @Description Creates a new user account with email verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body model.User true "User registration details"
// @Success 200 {object} object{status=string,message=string,data=response.AuthResponse} "User signed up successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 409 {object} utils.EmptyApiResponse "Username or email already exists"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /auth/signup [post]
func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}
	h.logger.Info("Received sign up request for user: ", user.Username)

	createdUser, err := h.authService.SignUp(&user, &h.ClientOtpMap)
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

// Login authenticates a user
// @Summary Login
// @Description Authenticates a user with username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body request.LoginRequest true "Login credentials"
// @Success 200 {object} object{status=string,message=string,data=response.AuthResponse,token=string} "Login successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 401 {object} utils.EmptyApiResponse "Invalid username or password"
// @Failure 404 {object} utils.EmptyApiResponse "User not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /auth [post]
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

// SendOTPEmail sends OTP to user's email
// @Summary Send OTP email
// @Description Sends a One-Time Password to the user's email for verification or password reset
// @Tags Authentication
// @Accept json
// @Produce json
// @Param otp body request.SendOTPEmailRequest true "OTP email request"
// @Success 200 {object} utils.EmptyApiResponse "OTP sent successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input or invalid purpose"
// @Failure 404 {object} utils.EmptyApiResponse "User not found"
// @Failure 409 {object} utils.EmptyApiResponse "Email already verified"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /auth/otp [post]
func (h *AuthHandler) SendOTPEmail(c *fiber.Ctx) error {
	var request request.SendOTPEmailRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if err := h.
		authService.
		GenerateAndSendOTPEmail(request.Email, request.Purpose, &h.ClientOtpMap); err != nil {
		if errors.Is(err, services.ErrInvalidPurpose) {
			return utils.HandleError(c, fiber.StatusBadRequest, "Invalid purpose", err)
		} else if errors.Is(err, services.ErrEmailAlreadyVerified) {
			return utils.HandleError(c, fiber.StatusConflict, "Email already verified", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	h.logger.Info("OTP sent to " + request.Email + " successfully.")
	return utils.HandleSuccess[any](c, "OTP sent successfully", nil)
}

// VerifyOTP verifies the OTP sent to user's email
// @Summary Verify OTP
// @Description Verifies the One-Time Password sent to the user's email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param otp body request.VerifyOTPRequest true "OTP verification request"
// @Success 200 {object} utils.EmptyApiResponse "OTP verified successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input or invalid OTP"
// @Failure 404 {object} utils.EmptyApiResponse "User not found or OTP not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /auth/verify [post]
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var request request.VerifyOTPRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	err := h.authService.VerifyOTP(request.Email, request.OTP, &h.ClientOtpMap)
	if err != nil {
		if errors.Is(err, services.ErrInvalidOTP) {
			return utils.HandleError(c, fiber.StatusBadRequest, "Invalid OTP", err)
		} else if errors.Is(err, services.ErrOTPNotFound) {
			return utils.HandleError(c, fiber.StatusNotFound, "OTP not found", err)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	h.logger.Println("OTP verified successfully for email", request.Email)
	return utils.HandleSuccess[any](c, "OTP verified successfully", nil)
}

// ResetPassword resets user's password
// @Summary Reset password
// @Description Resets the user's password after OTP verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Param reset body request.ResetPasswordRequest true "Password reset request"
// @Success 200 {object} utils.EmptyApiResponse "Password reset successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 404 {object} utils.EmptyApiResponse "User not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /auth/reset [post]
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

	return utils.HandleSuccess[any](c, "Password reset successfully", nil)
}

// GoogleLogin authenticates user via Google OAuth
// @Summary Google OAuth login
// @Description Authenticates a user using Google OAuth authorization code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param google body request.GoogleAuthRequest true "Google OAuth request"
// @Success 200 {object} object{status=string,message=string,data=response.AuthResponse,token=string} "Authenticated via Google successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Router /auth/google [post]
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

	h.logger.Info("User " + response.Username + " authenticated via Google OAuth")
	return utils.HandleLoginSuccess(c, "Authenticated via Google successfully", token, response)
}
