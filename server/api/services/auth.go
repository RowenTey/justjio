package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/RowenTey/JustJio/server/api/model"

	"github.com/golang-jwt/jwt"

	"golang.org/x/oauth2"
	googleOAuth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type AuthService struct {
	HashFunc      func(password string) (string, error)
	JwtSecret     string
	SendSMTPEmail func(from, to, subject, textBody string) error
	OAuthConfig   *oauth2.Config
	Logger        *log.Entry
}

// NOTE: used var instead of func to enable mocking in tests
var NewAuthService = func(
	hashFunc func(password string) (string, error),
	jwtSecret string,
	sendSMTPEmail func(from, to, subject, textBody string) error,
	oauthConfig *oauth2.Config,
) *AuthService {
	return &AuthService{
		HashFunc:      hashFunc,
		JwtSecret:     jwtSecret,
		SendSMTPEmail: sendSMTPEmail,
		OAuthConfig:   oauthConfig,
		Logger:        log.WithFields(log.Fields{"service": "AuthService"}),
	}
}

const TOKEN_EXPIRY_DURATION = time.Hour * 72 // 3 days

func (s *AuthService) SignUp(newUser *model.User) (*model.User, error) {
	var err error

	newUser.Password, err = s.HashFunc(newUser.Password)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *AuthService) CreateToken(user *model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["user_id"] = user.ID
	claims["user_email"] = user.Email
	claims["picture_url"] = user.PictureUrl
	claims["exp"] = time.Now().Add(TOKEN_EXPIRY_DURATION).Unix()

	t, err := token.SignedString([]byte(s.JwtSecret))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (s *AuthService) SendOTPEmail(otp, username, email, purpose string) error {
	from := config.Config("ADMIN_EMAIL")

	title := ""
	message := []byte("")
	if purpose == "verify-email" {
		title = "JustJio Email Verification"
		message = []byte("Welcome " + username + ",\r\n\r\n" +
			"We are happy to see you signed up with JustJio.\r\n\r\n" +
			"Your OTP is: " + otp)
	} else if purpose == "reset-password" {
		title = "JustJio Password Reset"
		message = []byte("Hi " + username + ",\r\n\r\n" +
			"Please use the following OTP to reset your password.\r\n\r\n" +
			"Your OTP is: " + otp)
	}

	err := s.SendSMTPEmail(from, email, title, string(message))
	if err != nil {
		return err
	}

	s.Logger.Info("OTP send to " + email + " successfully!")
	return nil
}

func (s *AuthService) VerifyOTP(storedOtp, email string, otp string) bool {
	return storedOtp == otp
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
	token, err := s.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	service, err := googleOAuth2.NewService(
		ctx, option.WithTokenSource(s.OAuthConfig.TokenSource(ctx, token)))
	if err != nil {
		return nil, err
	}

	userInfo, err := service.Userinfo.Get().Do()
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}
