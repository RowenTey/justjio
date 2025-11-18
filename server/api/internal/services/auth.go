package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/pkg/config"
	"github.com/RowenTey/JustJio/server/api/pkg/kafka"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
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
	kafkaClient   kafka.KafkaClient
	hashFunc      func(password string) (string, error)
	sendSMTPEmail func(from, to, subject, textBody string) error
	jwtSecret     string
	adminEmail    string
	oAuthConfig   *oauth2.Config
	logger        *logrus.Entry
}

func NewAuthService(
	userService *UserService,
	kafkaClient kafka.KafkaClient,
	hashFunc func(password string) (string, error),
	sendSMTPEmail func(from, to, subject, textBody string) error,
	conf *config.Config,
	logger *logrus.Logger,
) *AuthService {
	return &AuthService{
		userService:   userService,
		kafkaClient:   kafkaClient,
		hashFunc:      hashFunc,
		jwtSecret:     conf.JwtSecret,
		adminEmail:    conf.AdminEmail,
		sendSMTPEmail: sendSMTPEmail,
		oAuthConfig:   config.SetupGoogleOAuthConfig(conf),
		logger:        logger.WithFields(logrus.Fields{"service": "AuthService"}),
	}
}

func (as *AuthService) SignUp(ctx context.Context, newUser *models.User, otpMap *sync.Map) (*models.User, error) {
	var err error

	newUser.Password, err = as.hashFunc(newUser.Password)
	if err != nil {
		return nil, err
	}

	createdUser, err := as.userService.UpsertUser(ctx, newUser, true)
	if err != nil {
		return nil, err
	}

	// TODO: Make OTP have TTL
	// TODO: Should have some retry mechanism if email sending fails
	// Send OTP email
	go func() {
		otp := as.GenerateOTP()
		otpMap.Store(createdUser.Email, otp)
		as.logger.Info("Generated OTP for user: ", createdUser.Username)

		if err := as.
			SendOTPEmail(otp, createdUser.Username, createdUser.Email, VerifyEmailPurpose); err != nil {
			as.logger.Error("Error sending OTP email:", err)
			otpMap.Delete(createdUser.Email)
		}
		as.logger.Info("Sent OTP email to user: ", createdUser.Username)
	}()

	return createdUser, nil
}

func (as *AuthService) Login(ctx context.Context, username, password string) (string, *models.User, error) {
	start := time.Now()
	user, err := as.userService.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}
	as.logger.Infof("Time taken to find user: %v", time.Since(start))

	start = time.Now()
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", nil, ErrPasswordDoesNotMatch
	}
	as.logger.Infof("Time taken to check password: %v", time.Since(start))

	start = time.Now()
	token, err := as.CreateToken(user)
	if err != nil {
		return "", nil, err
	}
	as.logger.Infof("Time taken to create token: %v", time.Since(start))

	// TODO: Create on sign up instead of login?
	// Create user channel when login
	go func() {
		channel := fmt.Sprintf("user-%d", user.ID)
		if err := as.kafkaClient.CreateTopic(channel); err != nil {
			as.logger.Error("Error creating topic", err)
		}
	}()

	return token, user, nil
}

func (as *AuthService) GoogleLogin(ctx context.Context, code string) (string, *models.User, error) {
	googleUser, err := as.GetGoogleUser(code)
	if err != nil {
		return "", nil, err
	}

	user, err := as.userService.GetUserByEmail(ctx, googleUser.Email)
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

		newUser := &models.User{
			Username:     as.formatGoogleUsername(googleUser.Name),
			Email:        googleUser.Email,
			PictureUrl:   googleUser.Picture,
			Password:     hashedPassword,
			IsEmailValid: true,
		}
		user, err = as.userService.UpsertUser(ctx, newUser, true)
		if err != nil {
			return "", nil, err
		}
	}

	token, err := as.CreateToken(user)
	if err != nil {
		return "", nil, err
	}

	// Create user channel when login
	go func() {
		channel := fmt.Sprintf("user-%d", user.ID)
		if err := as.kafkaClient.CreateTopic(channel); err != nil {
			as.logger.Error("Error creating topic", err)
		}
	}()

	return token, user, nil
}

func (as *AuthService) CreateToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["user_id"] = user.ID
	claims["user_email"] = user.Email
	claims["picture_url"] = user.PictureUrl
	claims["exp"] = time.Now().Add(TOKEN_EXPIRY_DURATION).Unix()

	t, err := token.SignedString([]byte(as.jwtSecret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (as *AuthService) GenerateAndSendOTPEmail(ctx context.Context, email, purpose string, otpMap *sync.Map) error {
	if purpose != VerifyEmailPurpose && purpose != "reset-password" {
		return ErrInvalidPurpose
	}

	user, err := as.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if purpose == VerifyEmailPurpose && user.IsEmailValid {
		return ErrEmailAlreadyVerified
	}

	// TODO: Should have some retry mechanism if email sending fails
	otp := as.GenerateOTP()
	otpMap.Store(user.Email, otp)
	if err := as.SendOTPEmail(otp, user.Username, user.Email, purpose); err != nil {
		as.logger.Error("Error sending OTP email:", err)
		otpMap.Delete(user.Email)
	}

	return nil
}

func (as *AuthService) SendOTPEmail(otp, username, email, purpose string) error {
	from := as.adminEmail

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

	err := as.sendSMTPEmail(from, email, title, string(message))
	if err != nil {
		return err
	}

	as.logger.Info("OTP send to " + email + " successfully!")
	return nil
}

func (as *AuthService) VerifyOTP(ctx context.Context, email, otp string, otpMap *sync.Map) error {
	user, err := as.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	user.IsEmailValid = true
	if _, err := as.userService.UpsertUser(ctx, user, false); err != nil {
		return err
	}

	otpValue, exists := otpMap.Load(email)
	as.logger.Infof("Verifying OTP for email: %s, OTP exists: %t", email, exists)
	if !exists {
		return ErrOTPNotFound
	} else if otpValue != otp {
		return ErrInvalidOTP
	}

	// Delete OTP after verification
	otpMap.Delete(email)
	return nil
}

func (as *AuthService) ResetPassword(ctx context.Context, email, newPassword string) error {
	user, err := as.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	hashedPassword, err := as.hashFunc(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	if _, err := as.userService.UpsertUser(ctx, user, false); err != nil {
		return err
	}

	as.logger.Info("Password reset successfully for email ", email)
	return nil
}

func (as *AuthService) GenerateOTP() string {
	// Seed the random number generator
	rand.New(rand.NewSource(time.Now().Unix()))

	// Generate a random number between 000000 and 999999
	randomNumber := rand.Intn(1000000)

	// Format as a zero-padded 6-digit string
	return fmt.Sprintf("%06d", randomNumber)
}

func (as *AuthService) GetGoogleUser(code string) (*googleOAuth2.Userinfo, error) {
	ctx := context.Background()
	// Exchange code for token
	token, err := as.oAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	service, err := googleOAuth2.NewService(
		ctx, option.WithTokenSource(as.oAuthConfig.TokenSource(ctx, token)))
	if err != nil {
		return nil, err
	}

	userInfo, err := service.Userinfo.Get().Do()
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// formatGoogleUsername takes a name string and returns a formatted username
// by replacing spaces with underscores and adding today's date (DDMMYYYY)
func (as *AuthService) formatGoogleUsername(name string) string {
	// Replace spaces with underscores
	formattedName := strings.ReplaceAll(strings.ToLower(name), " ", "_")

	// Get current date and format it
	// Format: DDMMYY
	currentDate := time.Now().Format("020106")

	// Combine name and date
	return formattedName + "_" + currentDate
}
