package handlers

import (
	"errors"
	"fmt"
	"log"

	"github.com/RowenTey/JustJio/config"
	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model"
	"github.com/RowenTey/JustJio/model/request"
	"github.com/RowenTey/JustJio/model/response"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Global variable to access the ClientOTP map
// Store OTP with email as key
var ClientOTP = make(map[string]string)

func SignUp(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	authService := &services.AuthService{
		HashFunc:      util.HashPassword,
		SendSMTPEmail: util.SendSMTPEmail,
	}
	userService := &services.UserService{DB: database.DB}

	hashedPasswordUser, err := authService.SignUp(&user)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	createdUser, err := userService.CreateOrUpdateUser(hashedPasswordUser, true)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return util.HandleError(
				c, fiber.StatusConflict, "Username or email already exists", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	// Send OTP email
	go func() {
		otp := authService.GenerateOTP()
		ClientOTP[user.Email] = otp
		if err := authService.SendOTPEmail(otp, user.Username, user.Email, "verify-email"); err != nil {
			log.Println("Error sending OTP email:", err)
			delete(ClientOTP, user.Email)
		}
	}()

	response := response.AuthResponse{
		Email:      createdUser.Email,
		Username:   createdUser.Username,
		PictureUrl: createdUser.PictureUrl,
		UID:        createdUser.ID,
	}

	log.Println("[AUTH] User " + response.Username + " signed up successfully.")
	return util.HandleSuccess(c, "User signed up successfully", response)
}

func Login(c *fiber.Ctx, kafkaSvc *services.KafkaService) error {
	var input request.LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	username := input.Username
	user, err := (&services.UserService{DB: database.DB}).GetUserByUsername(username)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	if !util.CheckPasswordHash(input.Password, user.Password) {
		return util.HandleError(c, fiber.StatusUnauthorized, "Invalid password", errors.New("password does not match the user's password"))
	}

	authService := &services.AuthService{JwtSecret: config.Config("JWT_SECRET")}
	token, err := authService.CreateToken(user)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	// create user channel when login
	go func() {
		channel := fmt.Sprintf("user-%d", user.ID)
		if err := kafkaSvc.CreateTopic(channel); err != nil {
			log.Println("Error creating topic", err)
		}
	}()

	response := response.AuthResponse{
		Username:   user.Username,
		Email:      user.Email,
		PictureUrl: user.PictureUrl,
		UID:        user.ID,
	}

	log.Println("[AUTH] User " + response.Username + " logged in successfully.")
	return util.HandleLoginSuccess(c, "Login successfully", token, response)
}

func SendOTPEmail(c *fiber.Ctx) error {
	var request request.SendOTPEmailRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	if request.Purpose != "verify-email" && request.Purpose != "reset-password" {
		return util.HandleError(c, fiber.StatusBadRequest, "Invalid purpose", errors.New("invalid purpose"))
	}

	authService := &services.AuthService{SendSMTPEmail: util.SendSMTPEmail}
	userService := &services.UserService{DB: database.DB}

	user, err := userService.GetUserByEmail(request.Email)
	if err != nil {
		return util.HandleError(c, fiber.StatusNotFound, "Invalid email address", err)
	}

	if request.Purpose == "verify-email" && user.IsEmailValid {
		return util.HandleError(c, fiber.StatusConflict, "Email already verified", errors.New("email already verified"))
	}

	otp := authService.GenerateOTP()
	ClientOTP[user.Email] = otp
	if err := authService.SendOTPEmail(otp, user.Username, user.Email, request.Purpose); err != nil {
		log.Println("Error sending OTP email:", err)
		delete(ClientOTP, user.Email)
	}

	log.Println("[AUTH] OTP sent to " + request.Email + " successfully.")
	return util.HandleSuccess(c, "OTP sent successfully", nil)
}

func VerifyOTP(c *fiber.Ctx) error {
	var request request.VerifyOTPRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	authService := &services.AuthService{}
	userService := &services.UserService{DB: database.DB}

	user, err := userService.GetUserByEmail(request.Email)
	if err != nil {
		return util.HandleError(c, fiber.StatusNotFound, "Invalid email address", err)
	}

	user.IsEmailValid = true
	_, err = userService.CreateOrUpdateUser(user, false)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	otp, exists := ClientOTP[user.Email]
	if !exists {
		return util.HandleError(c, fiber.StatusBadRequest, "OTP not found", errors.New("OTP not found"))
	}

	isVerified := authService.VerifyOTP(otp, request.Email, request.OTP)
	if !isVerified {
		return util.HandleError(c, fiber.StatusBadRequest, "Invalid OTP", errors.New("invalid OTP"))
	}

	// Delete OTP after verification
	delete(ClientOTP, request.Email)

	log.Println("[AUTH] OTP verified successfully for email", request.Email)
	return util.HandleSuccess(c, "OTP verified successfully", nil)
}

func ResetPassword(c *fiber.Ctx) error {
	var request request.ResetPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	authService := &services.AuthService{HashFunc: util.HashPassword}
	userService := &services.UserService{DB: database.DB}

	user, err := userService.GetUserByEmail(request.Email)
	if err != nil {
		return util.HandleError(c, fiber.StatusNotFound, "Invalid email address", err)
	}

	hashedPassword, err := authService.HashFunc(request.Password)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	user.Password = hashedPassword
	_, err = userService.CreateOrUpdateUser(user, false)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	log.Println("[AUTH] Password reset successfully for email", request.Email)
	return util.HandleSuccess(c, "Password reset successfully", nil)
}

func GoogleLogin(c *fiber.Ctx) error {
	var request request.GoogleAuthRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	authService := &services.AuthService{
		JwtSecret:   config.Config("JWT_SECRET"),
		OAuthConfig: config.SetupGoogleOAuthConfig(),
	}
	userService := &services.UserService{DB: database.DB}

	googleUser, err := authService.GetGoogleUser(request.Code)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	user, err := userService.GetUserByEmail(googleUser.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return util.HandleInternalServerError(c, err)
	}

	// Create new user if not found
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Random password for OAuth user
		hashedPassword, err := util.HashPassword(util.GenerateRandomString(32))
		if err != nil {
			return util.HandleInternalServerError(c, err)
		}

		newUser := &model.User{
			Username:     util.FormatUsername(googleUser.Name),
			Email:        googleUser.Email,
			PictureUrl:   googleUser.Picture,
			Password:     hashedPassword,
			IsEmailValid: true,
		}
		user, err = userService.CreateOrUpdateUser(newUser, true)
		if err != nil {
			return util.HandleInternalServerError(c, err)
		}
	}

	token, err := authService.CreateToken(user)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	response := response.AuthResponse{
		Email:      user.Email,
		Username:   user.Username,
		PictureUrl: user.PictureUrl,
		UID:        user.ID,
	}

	log.Println("[AUTH] User " + response.Username + " authenticated via Google OAuth")
	return util.HandleLoginSuccess(c, "Authenticated via Google successfully", token, response)
}
