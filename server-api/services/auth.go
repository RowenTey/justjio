package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/RowenTey/JustJio/config"
	"github.com/RowenTey/JustJio/model"

	"github.com/golang-jwt/jwt"
)

type AuthService struct {
	HashFunc  func(password string) (string, error)
	JwtSecret string
	LoginAuth func(username, password string) smtp.Auth
	SendMail  func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
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
	claims["exp"] = time.Now().Add(TOKEN_EXPIRY_DURATION).Unix()

	t, err := token.SignedString([]byte(s.JwtSecret))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (s *AuthService) SendOTPEmail(ClientOTP *map[string]string, email string) error {
	otp := s.GenerateOTP()
	(*ClientOTP)[email] = otp

	// TODO: make env variable
	from := "justjio-team@outlook.com"
	password := config.Config("OUTLOOK_PASSWORD")
	to := []string{email}
	smtpHost := "smtp-mail.outlook.com"
	smtpPort := "587"

	// outlook requires LOGIN auth
	auth := s.LoginAuth(from, password)

	message := []byte("To: " + email + "\r\n" +
		"Subject: JustJio Email Verification\r\n" +
		"\r\n" +
		"Your OTP is: " + otp)

	err := s.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) VerifyOTP(ClientOTP *map[string]string, email string, otp string) error {
	if (*ClientOTP)[email] != otp {
		return errors.New("Invalid OTP")
	}

	delete(*ClientOTP, email)
	return nil
}

func (s *AuthService) GenerateOTP() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
	}
	return fmt.Sprintf("%x", b)
}
