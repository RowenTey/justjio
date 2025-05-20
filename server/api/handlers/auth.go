package handlers

import (
	"errors"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/model/response"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Global variable to access the ClientOTP map
// Store OTP with email as key
var ClientOTP sync.Map

var authLogger = log.WithFields(log.Fields{"service": "AuthHandler"})

func SignUp(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}
	authLogger.Info("Received sign up request for user: ", user.Username)

	authService := services.NewAuthService(
		utils.HashPassword,
		config.Config("JWT_SECRET"),
		utils.SendSMTPEmail,
		config.SetupGoogleOAuthConfig(),
	)
	userService := services.NewUserService(database.DB)

	hashedPasswordUser, err := authService.SignUp(&user)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	createdUser, err := userService.CreateOrUpdateUser(hashedPasswordUser, true)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return utils.HandleError(
				c, fiber.StatusConflict, "Username or email already exists", err)
		}
		return utils.HandleInternalServerError(c, err)
	}

	// Send OTP email
	go func() {
		otp := authService.GenerateOTP()
		ClientOTP.Store(user.Email, otp)
		authLogger.Info("Generated OTP for user: ", user.Username)

		if err := authService.SendOTPEmail(otp, user.Username, user.Email, "verify-email"); err != nil {
			authLogger.Error("Error sending OTP email:", err)
			ClientOTP.Delete(user.Email)
		}
		authLogger.Info("Sent OTP email to user: ", user.Username)
	}()

	response := response.AuthResponse{
		Email:      createdUser.Email,
		Username:   createdUser.Username,
		PictureUrl: createdUser.PictureUrl,
		UID:        createdUser.ID,
	}

	authLogger.Info("User " + response.Username + " signed up successfully.")
	return utils.HandleSuccess(c, "User signed up successfully", response)
}

func Login(c *fiber.Ctx, kafkaSvc *services.KafkaService) error {
	var input request.LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	username := input.Username
	user, err := services.NewUserService(database.DB).GetUserByUsername(username)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		return utils.HandleError(c, fiber.StatusUnauthorized, "Invalid password", errors.New("password does not match the user's password"))
	}

	authService := services.NewAuthService(
		utils.HashPassword,
		config.Config("JWT_SECRET"),
		utils.SendSMTPEmail,
		config.SetupGoogleOAuthConfig(),
	)
	token, err := authService.CreateToken(user)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	// create user channel when login
	go func() {
		channel := fmt.Sprintf("user-%d", user.ID)
		if err := kafkaSvc.CreateTopic(channel); err != nil {
			authLogger.Error("Error creating topic", err)
		}
	}()

	response := response.AuthResponse{
		Username:   user.Username,
		Email:      user.Email,
		PictureUrl: user.PictureUrl,
		UID:        user.ID,
	}

	authLogger.Info("User " + response.Username + " logged in successfully.")
	return utils.HandleLoginSuccess(c, "Login successfully", token, response)
}

func SendOTPEmail(c *fiber.Ctx) error {
	var request request.SendOTPEmailRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if request.Purpose != "verify-email" && request.Purpose != "reset-password" {
		return utils.HandleError(c, fiber.StatusBadRequest, "Invalid purpose", errors.New("invalid purpose"))
	}

	authService := services.NewAuthService(
		utils.HashPassword,
		config.Config("JWT_SECRET"),
		utils.SendSMTPEmail,
		config.SetupGoogleOAuthConfig(),
	)
	userService := services.NewUserService(database.DB)

	user, err := userService.GetUserByEmail(request.Email)
	if err != nil {
		return utils.HandleError(c, fiber.StatusNotFound, "Invalid email address", err)
	}

	if request.Purpose == "verify-email" && user.IsEmailValid {
		return utils.HandleError(c, fiber.StatusConflict, "Email already verified", errors.New("email already verified"))
	}

	otp := authService.GenerateOTP()
	ClientOTP.Store(user.Email, otp)
	if err := authService.SendOTPEmail(otp, user.Username, user.Email, request.Purpose); err != nil {
		authLogger.Error("Error sending OTP email:", err)
		ClientOTP.Delete(user.Email)
	}

	authLogger.Info("OTP sent to " + request.Email + " successfully.")
	return utils.HandleSuccess(c, "OTP sent successfully", nil)
}

func VerifyOTP(c *fiber.Ctx) error {
	var request request.VerifyOTPRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	authService := services.NewAuthService(
		utils.HashPassword,
		config.Config("JWT_SECRET"),
		utils.SendSMTPEmail,
		config.SetupGoogleOAuthConfig(),
	)
	userService := services.NewUserService(database.DB)

	user, err := userService.GetUserByEmail(request.Email)
	if err != nil {
		return utils.HandleError(c, fiber.StatusNotFound, "Invalid email address", err)
	}

	user.IsEmailValid = true
	_, err = userService.CreateOrUpdateUser(user, false)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	otpValue, exists := ClientOTP.Load(user.Email)
	if !exists {
		return utils.HandleError(c, fiber.StatusBadRequest, "OTP not found", errors.New("OTP not found"))
	}

	isVerified := authService.VerifyOTP(otpValue.(string), request.Email, request.OTP)
	if !isVerified {
		return utils.HandleError(c, fiber.StatusBadRequest, "Invalid OTP", errors.New("invalid OTP"))
	}

	// Delete OTP after verification
	ClientOTP.Delete(user.Email)

	authLogger.Println("OTP verified successfully for email", request.Email)
	return utils.HandleSuccess(c, "OTP verified successfully", nil)
}

func ResetPassword(c *fiber.Ctx) error {
	var request request.ResetPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	authService := services.NewAuthService(
		utils.HashPassword,
		config.Config("JWT_SECRET"),
		utils.SendSMTPEmail,
		config.SetupGoogleOAuthConfig(),
	)
	userService := services.NewUserService(database.DB)

	user, err := userService.GetUserByEmail(request.Email)
	if err != nil {
		return utils.HandleError(c, fiber.StatusNotFound, "Invalid email address", err)
	}

	hashedPassword, err := authService.HashFunc(request.Password)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	user.Password = hashedPassword
	_, err = userService.CreateOrUpdateUser(user, false)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	authLogger.Info("Password reset successfully for email ", request.Email)
	return utils.HandleSuccess(c, "Password reset successfully", nil)
}

func GoogleLogin(c *fiber.Ctx, kafkaSvc *services.KafkaService) error {
	var request request.GoogleAuthRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	authService := services.NewAuthService(
		utils.HashPassword,
		config.Config("JWT_SECRET"),
		utils.SendSMTPEmail,
		config.SetupGoogleOAuthConfig(),
	)
	userService := services.NewUserService(database.DB)

	googleUser, err := authService.GetGoogleUser(request.Code)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	user, err := userService.GetUserByEmail(googleUser.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.HandleInternalServerError(c, err)
	}

	// Create new user if not found
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Random password for OAuth user
		hashedPassword, err := utils.HashPassword(utils.GenerateRandomString(32))
		if err != nil {
			return utils.HandleInternalServerError(c, err)
		}

		newUser := &model.User{
			Username:     utils.FormatUsername(googleUser.Name),
			Email:        googleUser.Email,
			PictureUrl:   googleUser.Picture,
			Password:     hashedPassword,
			IsEmailValid: true,
		}
		user, err = userService.CreateOrUpdateUser(newUser, true)
		if err != nil {
			return utils.HandleInternalServerError(c, err)
		}
	}

	token, err := authService.CreateToken(user)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	// create user channel when login
	go func() {
		channel := fmt.Sprintf("user-%d", user.ID)
		if err := kafkaSvc.CreateTopic(channel); err != nil {
			authLogger.Error("Error creating topic", err)
		}
	}()

	response := response.AuthResponse{
		Email:      user.Email,
		Username:   user.Username,
		PictureUrl: user.PictureUrl,
		UID:        user.ID,
	}

	authLogger.Println("User " + response.Username + " authenticated via Google OAuth")
	return utils.HandleLoginSuccess(c, "Authenticated via Google successfully", token, response)
}
