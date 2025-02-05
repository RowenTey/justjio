package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"

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
		HashFunc:  util.HashPassword,
		LoginAuth: util.NewLoginAuth,
		SendMail:  smtp.SendMail,
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

	if err := authService.SendOTPEmail(&ClientOTP, user.Email); err != nil {
		log.Println("Error sending OTP email:", err)
		delete(ClientOTP, user.Email)
	}

	response := response.AuthResponse{
		Email:    createdUser.Email,
		Username: createdUser.Username,
		UID:      createdUser.ID,
	}

	log.Println("User " + response.Username + " signed up successfully.")
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
		return util.HandleError(c, fiber.StatusUnauthorized, "Invalid password", errors.New("Password does not match the user's password"))
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
		Username: user.Username,
		Email:    user.Email,
		UID:      user.ID,
	}

	log.Println("User " + response.Username + " logged in successfully.")
	return util.HandleLoginSuccess(c, "Login successfully", token, response)
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

	err = authService.VerifyOTP(&ClientOTP, request.Email, request.OTP)
	if err != nil {
		return util.HandleError(c, fiber.StatusBadRequest, "Invalid OTP", err)
	}

	log.Println("OTP verified successfully for email", request.Email)
	return util.HandleSuccess(c, "OTP verified successfully", nil)
}
