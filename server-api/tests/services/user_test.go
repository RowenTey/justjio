package test_services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/model"
	"github.com/RowenTey/JustJio/services"
)

type UserServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	userService *services.UserService
}

func (s *UserServiceTestSuite) SetupTest() {
	var (
		db  *sql.DB
		err error
	)

	db, s.mock, err = sqlmock.New()
	assert.NoError(s.T(), err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})
	s.DB, err = gorm.Open(dialector, &gorm.Config{})
	assert.NoError(s.T(), err)

	s.userService = &services.UserService{DB: s.DB}
}

func (s *UserServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *UserServiceTestSuite) TestGetUserByID_Success() {
	// arrange
	expectedUser := &model.User{
		ID:           1,
		Username:     "johndoe",
		Email:        "john@example.com",
		Password:     "hashedpassword",
		Name:         "John Doe",
		PhoneNum:     "1234567890",
		IsEmailValid: true,
		IsOnline:     false,
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
		UpdatedAt:    time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "name", "phone_num",
		"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
		AddRow(
			expectedUser.ID,
			expectedUser.Username,
			expectedUser.Email,
			expectedUser.Password,
			expectedUser.Name,
			expectedUser.PhoneNum,
			expectedUser.IsEmailValid,
			expectedUser.IsOnline,
			expectedUser.LastSeen,
			expectedUser.RegisteredAt,
			expectedUser.UpdatedAt,
		)

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnRows(rows)

	// act
	user, err := s.userService.GetUserByID("1")

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedUser.ID, user.ID)
	assert.Equal(s.T(), expectedUser.Username, user.Username)
	assert.Equal(s.T(), expectedUser.Email, user.Email)
	assert.Equal(s.T(), expectedUser.Name, user.Name)
	assert.Equal(s.T(), expectedUser.PhoneNum, user.PhoneNum)
	assert.Equal(s.T(), expectedUser.IsEmailValid, user.IsEmailValid)
	assert.Equal(s.T(), expectedUser.IsOnline, user.IsOnline)
	assert.Equal(s.T(), expectedUser.LastSeen.Unix(), user.LastSeen.Unix())
	assert.Equal(s.T(), expectedUser.RegisteredAt.Unix(), user.RegisteredAt.Unix())
}

func (s *UserServiceTestSuite) TestGetUserByID_NotFound() {
	// arrange
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	user, err := s.userService.GetUserByID("1")

	// assert
	assert.Error(s.T(), err)
	assert.True(s.T(), err == gorm.ErrRecordNotFound)
	assert.Nil(s.T(), user)
}

func (s *UserServiceTestSuite) TestUpdateUserField_Success() {
	// arrange
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "name", "phone_num", "is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(1, "johndoe", "john@example.com", "hashedpassword", "John Doe", "1234567890", true, false, time.Now(), time.Now(), time.Now()))

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET "username"=\$1,"email"=\$2,"password"=\$3,"name"=\$4,"phone_num"=\$5,"is_email_valid"=\$6,"is_online"=\$7,"last_seen"=\$8,"registered_at"=\$9,"updated_at"=\$10 WHERE "id" = \$11`).
		WithArgs(
			"johndoe",
			"john@example.com",
			"hashedpassword",
			"Jane Doe",
			"1234567890",
			true,
			false,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			1,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.UpdateUserField("1", "name", "Jane Doe")

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestDeleteUser_Success() {
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "users" WHERE "users"."id" = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err := s.userService.DeleteUser("1")

	assert.NoError(s.T(), err)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
