package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/golang-jwt/jwt"

	"golang.org/x/oauth2"
	googleOAuth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var (
	ErrPasswordDoesNotMatch = errors.New("password does not match the user's password")
	ErrInvalidPurpose       = errors.New("invalid purpose for OTP generation")
	ErrEmailAlreadyVerified = errors.New("email already verified")
	ErrOTPNotFound          = errors.New("OTP not found")
	ErrInvalidOTP           = errors.New("invalid OTP")
	VerifyEmailPurpose      = "verify-email"
)

const (
	TOKEN_EXPIRY_DURATION = time.Hour * 72 // 3 days
)

type AuthService struct {
	userService   *UserService
	kafkaService  KafkaService
	hashFunc      func(password string) (string, error)
	sendSMTPEmail func(from, to, subject, textBody string) error
	jwtSecret     string
	adminEmail    string
	oAuthConfig   *oauth2.Config
	logger        *logrus.Entry
}

func NewAuthService(
	userService *UserService,
	kafkaService KafkaService,
	hashFunc func(password string) (string, error),
	sendSMTPEmail func(from, to, subject, textBody string) error,
	conf *config.Config,
	logger *logrus.Logger,
) *AuthService {
	return &AuthService{
		userService:   userService,
		kafkaService:  kafkaService,
		hashFunc:      hashFunc,
		jwtSecret:     conf.JwtSecret,
		adminEmail:    conf.AdminEmail,
		sendSMTPEmail: sendSMTPEmail,
		oAuthConfig:   config.SetupGoogleOAuthConfig(conf),
		logger:        logger.WithFields(logrus.Fields{"service": "AuthService"}),
	}
}

func (s *AuthService) SignUp(newUser *model.User, otpMap *sync.Map) (*model.User, error) {
	var err error

	newUser.Password, err = s.hashFunc(newUser.Password)
	if err != nil {
		return nil, err
	}

	createdUser, err := s.userService.CreateOrUpdateUser(newUser, true)
	if err != nil {
		return nil, err
	}

	// TODO: Make OTP have TTL
	// Send OTP email
	go func() {
		otp := s.GenerateOTP()
		otpMap.Store(createdUser.Email, otp)
		s.logger.Info("Generated OTP for user: ", createdUser.Username)

		if err := s.
			SendOTPEmail(otp, createdUser.Username, createdUser.Email, VerifyEmailPurpose); err != nil {
			s.logger.Error("Error sending OTP email:", err)
			otpMap.Delete(createdUser.Email)
		}
		s.logger.Info("Sent OTP email to user: ", createdUser.Username)
	}()

	return createdUser, nil
}

func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	user, err := s.userService.GetUserByUsername(username)
	if err != nil {
		return "", nil, err
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", nil, ErrPasswordDoesNotMatch
	}

	token, err := s.CreateToken(user)
	if err != nil {
		return "", nil, err
	}

	// TODO: Create on sign up instead of login?
	// create user channel when login
	go func() {
		channel := fmt.Sprintf("user-%d", user.ID)
		if err := s.kafkaService.CreateTopic(channel); err != nil {
			s.logger.Error("Error creating topic", err)
		}
	}()

	return token, user, nil
}

func (s *AuthService) GoogleLogin(code string) (string, *model.User, error) {
	googleUser, err := s.GetGoogleUser(code)
	if err != nil {
		return "", nil, err
	}

	user, err := s.userService.GetUserByEmail(googleUser.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil, err
	}

	// Create new user if not found
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Random password for OAuth user
		hashedPassword, err := utils.HashPassword(utils.GenerateRandomString(32))
		if err != nil {
			return "", nil, err
		}

		newUser := &model.User{
			Username:     utils.FormatUsername(googleUser.Name),
			Email:        googleUser.Email,
			PictureUrl:   googleUser.Picture,
			Password:     hashedPassword,
			IsEmailValid: true,
		}
		user, err = s.userService.CreateOrUpdateUser(newUser, true)
		if err != nil {
			return "", nil, err
		}
	}

	token, err := s.CreateToken(user)
	if err != nil {
		return "", nil, err
	}

	// create user channel when login
	go func() {
		channel := fmt.Sprintf("user-%d", user.ID)
		if err := s.kafkaService.CreateTopic(channel); err != nil {
			s.logger.Error("Error creating topic", err)
		}
	}()

	return token, user, nil
}

func (s *AuthService) CreateToken(user *model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["user_id"] = user.ID
	claims["user_email"] = user.Email
	claims["picture_url"] = user.PictureUrl
	claims["exp"] = time.Now().Add(TOKEN_EXPIRY_DURATION).Unix()

	t, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (s *AuthService) GenerateAndSendOTPEmail(email, purpose string, otpMap *sync.Map) error {
	if purpose != VerifyEmailPurpose && purpose != "reset-password" {
		return ErrInvalidPurpose
	}

	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		return err
	}

	if purpose == VerifyEmailPurpose && user.IsEmailValid {
		return ErrEmailAlreadyVerified
	}

	// TODO: Should have some retry mechanism if email sending fails
	otp := s.GenerateOTP()
	otpMap.Store(user.Email, otp)
	if err := s.SendOTPEmail(otp, user.Username, user.Email, purpose); err != nil {
		s.logger.Error("Error sending OTP email:", err)
		otpMap.Delete(user.Email)
	}

	return nil
}

func (s *AuthService) SendOTPEmail(otp, username, email, purpose string) error {
	from := s.adminEmail

	title := ""
	message := []byte("")
	switch purpose {
	case VerifyEmailPurpose:
		title = "JustJio Email Verification"
		message = []byte("Welcome " + username + ",\r\n\r\n" +
			"We are happy to see you signed up with JustJio.\r\n\r\n" +
			"Your OTP is: " + otp)
	case "reset-password":
		title = "JustJio Password Reset"
		message = []byte("Hi " + username + ",\r\n\r\n" +
			"Please use the following OTP to reset your password.\r\n\r\n" +
			"Your OTP is: " + otp)
	}

	err := s.sendSMTPEmail(from, email, title, string(message))
	if err != nil {
		return err
	}

	s.logger.Info("OTP send to " + email + " successfully!")
	return nil
}

func (s *AuthService) VerifyOTP(email, otp string, otpMap *sync.Map) error {
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		return err
	}

	user.IsEmailValid = true
	if _, err := s.userService.CreateOrUpdateUser(user, false); err != nil {
		return err
	}

	otpValue, exists := otpMap.Load(email)
	s.logger.Infof("Verifying OTP for email: %s, OTP exists: %t", email, exists)
	if !exists {
		return ErrOTPNotFound
	} else if otpValue != otp {
		return ErrInvalidOTP
	}

	// Delete OTP after verification
	otpMap.Delete(email)
	return nil
}

func (s *AuthService) ResetPassword(email, newPassword string) error {
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		return err
	}

	hashedPassword, err := s.hashFunc(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	if _, err := s.userService.CreateOrUpdateUser(user, false); err != nil {
		return err
	}

	s.logger.Info("Password reset successfully for email ", email)
	return nil
}

func (s *AuthService) GenerateOTP() string {
	// Seed the random number generator
	rand.New(rand.NewSource(time.Now().Unix()))

	// Generate a random number between 000000 and 999999
	randomNumber := rand.Intn(1000000)

	// Format as a zero-padded 6-digit string
	return fmt.Sprintf("%06d", randomNumber)
}

func (s *AuthService) GetGoogleUser(code string) (*googleOAuth2.Userinfo, error) {
	ctx := context.Background()
	// Exchange code for token
	token, err := s.oAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	service, err := googleOAuth2.NewService(
		ctx, option.WithTokenSource(s.oAuthConfig.TokenSource(ctx, token)))
	if err != nil {
		return nil, err
	}

	userInfo, err := service.Userinfo.Get().Do()
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}
