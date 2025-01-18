package test_services

import (
	"net/smtp"
	"testing"

	"github.com/RowenTey/JustJio/model"
	"github.com/RowenTey/JustJio/services"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.Suite
	authService *services.AuthService
}

func (s *AuthServiceTestSuite) SetupTest() {
	s.authService = &services.AuthService{}
}

func (s *AuthServiceTestSuite) TestSignUp_Success() {
	// arrange
	s.authService.HashFunc = func(password string) (string, error) {
		return "hashedpassword", nil
	}

	user := &model.User{
		Username: "testuser",
		Password: "testpassword",
	}

	// act
	newUser, err := s.authService.SignUp(user)

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "hashedpassword", newUser.Password)
}

func (s *AuthServiceTestSuite) TestCreateToken_Success() {
	user := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "testuser@example.com",
	}
	s.authService.JwtSecret = "testsecret"

	token, createTokenErr := s.authService.CreateToken(user)
	parsedToken, validateTokenErr := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("testsecret"), nil
	})

	assert.NoError(s.T(), createTokenErr)
	assert.NoError(s.T(), validateTokenErr)
	assert.True(s.T(), parsedToken.Valid)
}

func (s *AuthServiceTestSuite) TestSendOTPEmail_Success() {
	s.authService.LoginAuth = func(username, password string) smtp.Auth {
		return nil
	}
	s.authService.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return nil
	}
	clientOTP := make(map[string]string)

	err := s.authService.SendOTPEmail(&clientOTP, "test@example.com")

	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), clientOTP["test@example.com"])
}

func (s *AuthServiceTestSuite) TestVerifyOTP_Success() {
	clientOTP := map[string]string{
		"test@example.com": "123456",
	}

	err := s.authService.VerifyOTP(&clientOTP, "test@example.com", "123456")
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), clientOTP["test@example.com"])
}

func (s *AuthServiceTestSuite) TestVerifyOTP_Failure() {
	clientOTP := map[string]string{
		"test@example.com": "123456",
	}

	err := s.authService.VerifyOTP(&clientOTP, "test@example.com", "654321")
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "Invalid OTP", err.Error())
	assert.NotEmpty(s.T(), clientOTP["test@example.com"])
}

func (s *AuthServiceTestSuite) TestGenerateOTP_Success() {
	otp := s.authService.GenerateOTP()
	assert.NotEmpty(s.T(), otp)
	assert.Len(s.T(), otp, 12) // Since the OTP is generated with 6 bytes and converted to hex
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
