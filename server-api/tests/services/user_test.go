package test_services

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

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

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
		},
	)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})
	s.DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
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
		IsEmailValid: true,
		IsOnline:     false,
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
		UpdatedAt:    time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password",
		"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
		AddRow(
			expectedUser.ID,
			expectedUser.Username,
			expectedUser.Email,
			expectedUser.Password,
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

func (s *UserServiceTestSuite) TestUpdateUserField_Username() {
	// arrange
	now := time.Now()
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "oldjohndoe", "john@example.com", "hashedpassword",
				"https://default-image.jpg", true, false, now, now, now))

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs("newjohndoe", "john@example.com", "hashedpassword",
			"https://default-image.jpg", true, false, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.UpdateUserField("1", "username", "newjohndoe")

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestUpdateUserField_IsEmailValid() {
	// arrange
	now := time.Now()
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "oldjohndoe", "john@example.com", "hashedpassword",
				"https://default-image.jpg", false, false, now, now, now))

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs("oldjohndoe", "john@example.com", "hashedpassword",
			"https://default-image.jpg", true, false, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.UpdateUserField("1", "isEmailValid", true)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestUpdateUserField_IsOnline() {
	// arrange
	now := time.Now()
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "oldjohndoe", "john@example.com", "hashedpassword",
				"https://default-image.jpg", true, false, now, now, now))

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs("oldjohndoe", "john@example.com", "hashedpassword",
			"https://default-image.jpg", true, true, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.UpdateUserField("1", "isOnline", true)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestUpdateUserField_LastSeen() {
	// arrange
	now := time.Now()
	newTime := time.Now().Add(time.Hour)

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "oldjohndoe", "john@example.com", "hashedpassword",
				"https://default-image.jpg", true, false, now, now, now))

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs("oldjohndoe", "john@example.com", "hashedpassword",
			"https://default-image.jpg", true, false, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.UpdateUserField("1", "lastSeen", newTime)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestUpdateUserField_UserNotFound() {
	// arrange
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("999", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	err := s.userService.UpdateUserField("999", "username", "johndoe")

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}

func (s *UserServiceTestSuite) TestUpdateUserField_UnsupportedField() {
	// arrange
	now := time.Now()
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "johndoe", "john@example.com", "hashedpassword",
				true, false, now, now, now))

	// act
	err := s.userService.UpdateUserField("1", "unsupportedField", "value")

	// assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not supported for update")
}

func (s *UserServiceTestSuite) TestUpdateUserField_DatabaseError() {
	// arrange
	now := time.Now()
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "oldjohndoe", "john@example.com", "hashedpassword",
				"https://default-image.jpg", true, false, now, now, now))

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs("newjohndoe", "john@example.com", "hashedpassword",
			"https://default-image.jpg", true, false, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	err := s.userService.UpdateUserField("1", "username", "newjohndoe")

	// assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
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

func (s *UserServiceTestSuite) TestSendFriendRequest_Success() {
	// arrange
	senderID := uint(1)
	receiverID := uint(2)

	// Mock preloading Friends for the sender
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(senderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(senderID, "sender", "sender@example.com", "hashedpw",
				"", true, false, time.Now(), time.Now(), time.Now()))

	// Mock loading Friends association
	s.mock.ExpectQuery(`SELECT \* FROM "user_friends" WHERE "user_friends"."user_id" = \$1`).
		WithArgs(senderID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "friend_id"}))

	// Mock checking for existing friend requests
	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE \(\(sender_id = \$1 AND receiver_id = \$2\) OR \(sender_id = \$3 AND receiver_id = \$4\)\) AND status = \$5 ORDER BY "friend_requests"."id" LIMIT \$6`).
		WithArgs(senderID, receiverID, receiverID, senderID, "pending", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock creating the friend request
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "friend_requests" .*`).
		WithArgs(senderID, receiverID, "pending", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.SendFriendRequest(senderID, receiverID)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestSendFriendRequest_SameUser() {
	// arrange
	senderID := uint(1)

	// act
	err := s.userService.SendFriendRequest(senderID, senderID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "cannot send friend request to yourself", err.Error())
}

func (s *UserServiceTestSuite) TestSendFriendRequest_SenderNotFound() {
	// arrange
	senderID := uint(1)
	receiverID := uint(2)

	// Mock preloading Friends for the sender with error
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(senderID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	err := s.userService.SendFriendRequest(senderID, receiverID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "sender not found", err.Error())
}

func (s *UserServiceTestSuite) TestSendFriendRequest_AlreadyFriends() {
	// arrange
	senderID := uint(1)
	receiverID := uint(2)

	// Mock preloading Friends for the sender
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(senderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(senderID, "sender", "sender@example.com", "hashedpw",
				"", true, false, time.Now(), time.Now(), time.Now()))

	// For many-to-many relationships, GORM first queries the join table
	s.mock.ExpectQuery(`SELECT \* FROM "user_friends" WHERE "user_friends"."user_id" = \$1`).
		WithArgs(senderID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "friend_id"}).
			AddRow(senderID, receiverID))

	// Then it fetches the associated records from the target table
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1`).
		WithArgs(receiverID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(receiverID, "receiver", "receiver@example.com", "hashedpw",
				"", true, false, time.Now(), time.Now(), time.Now()))

	// act
	err := s.userService.SendFriendRequest(senderID, receiverID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "already friends", err.Error())
}

func (s *UserServiceTestSuite) TestSendFriendRequest_RequestAlreadySent() {
	// arrange
	senderID := uint(1)
	receiverID := uint(2)

	// Mock preloading Friends for the sender
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(senderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(senderID, "sender", "sender@example.com", "hashedpw",
				"", true, false, time.Now(), time.Now(), time.Now()))

	// Mock loading Friends association
	s.mock.ExpectQuery(`SELECT \* FROM "user_friends" WHERE "user_friends"."user_id" = \$1`).
		WithArgs(senderID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "friend_id"}))

	// Mock finding an existing friend request
	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE \(\(sender_id = \$1 AND receiver_id = \$2\) OR \(sender_id = \$3 AND receiver_id = \$4\)\) AND status = \$5 ORDER BY "friend_requests"."id" LIMIT \$6`).
		WithArgs(senderID, receiverID, receiverID, senderID, "pending", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "status"}).
			AddRow(1, senderID, receiverID, "pending"))

	// act
	err := s.userService.SendFriendRequest(senderID, receiverID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "friend request already sent", err.Error())
}

func (s *UserServiceTestSuite) TestSendFriendRequest_DatabaseError() {
	// arrange
	senderID := uint(1)
	receiverID := uint(2)

	// Mock preloading Friends for the sender
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(senderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(senderID, "sender", "sender@example.com", "hashedpw",
				"", true, false, time.Now(), time.Now(), time.Now()))

	// Mock loading Friends association
	s.mock.ExpectQuery(`SELECT \* FROM "user_friends" WHERE "user_friends"."user_id" = \$1`).
		WithArgs(senderID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "friend_id"}))

	// Mock checking for existing friend requests
	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE \(\(sender_id = \$1 AND receiver_id = \$2\) OR \(sender_id = \$3 AND receiver_id = \$4\)\) AND status = \$5 ORDER BY "friend_requests"."id" LIMIT \$6`).
		WithArgs(senderID, receiverID, receiverID, senderID, "pending", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock database error when creating request
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "friend_requests" .*`).
		WithArgs(senderID, receiverID, "pending", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database connection error"))
	s.mock.ExpectRollback()

	// act
	err := s.userService.SendFriendRequest(senderID, receiverID)

	// assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database connection error")
}

func (s *UserServiceTestSuite) TestGetUserByUsername_Success() {
	// arrange
	expectedUser := &model.User{
		ID:           1,
		Username:     "johndoe",
		Email:        "john@example.com",
		Password:     "hashedpassword",
		PictureUrl:   "https://default-image.jpg",
		IsEmailValid: true,
		IsOnline:     false,
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
		UpdatedAt:    time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "picture_url",
		"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
		AddRow(
			expectedUser.ID,
			expectedUser.Username,
			expectedUser.Email,
			expectedUser.Password,
			expectedUser.PictureUrl,
			expectedUser.IsEmailValid,
			expectedUser.IsOnline,
			expectedUser.LastSeen,
			expectedUser.RegisteredAt,
			expectedUser.UpdatedAt,
		)

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("johndoe", 1).
		WillReturnRows(rows)

	// act
	user, err := s.userService.GetUserByUsername("johndoe")

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedUser.ID, user.ID)
	assert.Equal(s.T(), expectedUser.Username, user.Username)
	assert.Equal(s.T(), expectedUser.Email, user.Email)
	assert.Equal(s.T(), expectedUser.PictureUrl, user.PictureUrl)
	assert.Equal(s.T(), expectedUser.IsEmailValid, user.IsEmailValid)
	assert.Equal(s.T(), expectedUser.IsOnline, user.IsOnline)
}

func (s *UserServiceTestSuite) TestGetUserByUsername_NotFound() {
	// arrange
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("nonexistentuser", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	user, err := s.userService.GetUserByUsername("nonexistentuser")

	// assert
	assert.Error(s.T(), err)
	assert.True(s.T(), err == gorm.ErrRecordNotFound)
	assert.Nil(s.T(), user)
}

func (s *UserServiceTestSuite) TestGetUserByEmail_Success() {
	// arrange
	expectedUser := &model.User{
		ID:           1,
		Username:     "johndoe",
		Email:        "john@example.com",
		Password:     "hashedpassword",
		PictureUrl:   "https://default-image.jpg",
		IsEmailValid: true,
		IsOnline:     false,
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
		UpdatedAt:    time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "picture_url",
		"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
		AddRow(
			expectedUser.ID,
			expectedUser.Username,
			expectedUser.Email,
			expectedUser.Password,
			expectedUser.PictureUrl,
			expectedUser.IsEmailValid,
			expectedUser.IsOnline,
			expectedUser.LastSeen,
			expectedUser.RegisteredAt,
			expectedUser.UpdatedAt,
		)

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("john@example.com", 1).
		WillReturnRows(rows)

	// act
	user, err := s.userService.GetUserByEmail("john@example.com")

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedUser.ID, user.ID)
	assert.Equal(s.T(), expectedUser.Username, user.Username)
	assert.Equal(s.T(), expectedUser.Email, user.Email)
}

func (s *UserServiceTestSuite) TestGetUserByEmail_NotFound() {
	// arrange
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("nonexistent@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	user, err := s.userService.GetUserByEmail("nonexistent@example.com")

	// assert
	assert.Error(s.T(), err)
	assert.True(s.T(), err == gorm.ErrRecordNotFound)
	assert.Nil(s.T(), user)
}

func (s *UserServiceTestSuite) TestCreateOrUpdateUser_Create() {
	// arrange
	now := time.Now()
	user := &model.User{
		Username:   "newuser",
		Email:      "new@example.com",
		Password:   "hashedpassword",
		PictureUrl: "https://default-image.jpg",
	}

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(
			user.Username,
			user.Email,
			user.Password,
			user.PictureUrl,
			user.IsEmailValid,
			user.IsOnline,
			sqlmock.AnyArg(), // LastSeen
			sqlmock.AnyArg(), // RegisteredAt
			sqlmock.AnyArg(), // UpdatedAt
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	s.mock.ExpectCommit()

	// act
	result, err := s.userService.CreateOrUpdateUser(user, true)

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), uint(1), result.ID)
	assert.Equal(s.T(), user.Username, result.Username)
	assert.Equal(s.T(), user.Email, result.Email)
	assert.WithinDuration(s.T(), now, result.RegisteredAt, time.Second*10)
	assert.WithinDuration(s.T(), now, result.UpdatedAt, time.Second*10)
}

func (s *UserServiceTestSuite) TestCreateOrUpdateUser_Update() {
	// arrange
	now := time.Now()
	user := &model.User{
		ID:         1,
		Username:   "existinguser",
		Email:      "existing@example.com",
		Password:   "hashedpassword",
		PictureUrl: "https://default-image.jpg",
		UpdatedAt:  time.Now().Add(-24 * time.Hour), // Old timestamp
	}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs(
			user.Username,
			user.Email,
			user.Password,
			user.PictureUrl,
			user.IsEmailValid,
			user.IsOnline,
			sqlmock.AnyArg(), // LastSeen
			sqlmock.AnyArg(), // RegisteredAt (not updated)
			sqlmock.AnyArg(), // UpdatedAt (updated)
			user.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	result, err := s.userService.CreateOrUpdateUser(user, false)

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), user.ID, result.ID)
	assert.Equal(s.T(), user.Username, result.Username)
	assert.Equal(s.T(), user.Email, result.Email)
	assert.WithinDuration(s.T(), now, result.UpdatedAt, time.Second*10)
}

func (s *UserServiceTestSuite) TestCreateOrUpdateUser_Error() {
	// arrange
	user := &model.User{
		Username:   "erroruser",
		Email:      "error@example.com",
		Password:   "hashedpassword",
		PictureUrl: "https://default-image.jpg",
	}

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(
			user.Username,
			user.Email,
			user.Password,
			user.PictureUrl,
			user.IsEmailValid,
			user.IsOnline,
			sqlmock.AnyArg(), // LastSeen
			sqlmock.AnyArg(), // RegisteredAt
			sqlmock.AnyArg(), // UpdatedAt
		).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	result, err := s.userService.CreateOrUpdateUser(user, true)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *UserServiceTestSuite) TestValidateUsers_Success() {
	// arrange
	userIds := []string{"1", "2", "3"}
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "picture_url",
		"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"})

	for _, id := range userIds {
		idInt, _ := strconv.ParseUint(id, 10, 64)
		rows.AddRow(
			idInt,
			"user"+id,
			"user"+id+"@example.com",
			"hashedpassword",
			"https://default-image.jpg",
			true,
			false,
			now,
			now,
			now,
		)
	}

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2,\$3\)`).
		WithArgs("1", "2", "3").
		WillReturnRows(rows)

	// act
	users, err := s.userService.ValidateUsers(userIds)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), users)
	assert.Equal(s.T(), 3, len(*users))
	for i, id := range userIds {
		idInt, _ := strconv.ParseUint(id, 10, 64)
		assert.Equal(s.T(), uint(idInt), (*users)[i].ID)
		assert.Equal(s.T(), "user"+id, (*users)[i].Username)
	}
}

func (s *UserServiceTestSuite) TestValidateUsers_EmptyInput() {
	// arrange
	userIds := []string{}

	// act
	users, err := s.userService.ValidateUsers(userIds)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), users)
	assert.Equal(s.T(), 0, len(*users))
}

func (s *UserServiceTestSuite) TestValidateUsers_DatabaseError() {
	// arrange
	userIds := []string{"1", "2", "3"}

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2,\$3\)`).
		WithArgs("1", "2", "3").
		WillReturnError(errors.New("database error"))

	// act
	users, err := s.userService.ValidateUsers(userIds)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), users)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *UserServiceTestSuite) TestMarkOnline_Success() {
	// arrange
	userId := "1"
	now := time.Now()

	// First, the user is fetched
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userId, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "johndoe", "john@example.com", "hashedpassword",
				"https://default-image.jpg", true, false, now, now, now))

	// Then, the user is updated
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs("johndoe", "john@example.com", "hashedpassword",
			"https://default-image.jpg", true, true, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.MarkOnline(userId)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestMarkOnline_UserNotFound() {
	// arrange
	userId := "999"

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userId, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	err := s.userService.MarkOnline(userId)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}

func (s *UserServiceTestSuite) TestMarkOffline_Success() {
	// arrange
	userId := "1"
	now := time.Now()

	// First query to update isOnline
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userId, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "johndoe", "john@example.com", "hashedpassword",
				"https://default-image.jpg", true, true, now, now, now))

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs("johndoe", "john@example.com", "hashedpassword",
			"https://default-image.jpg", true, false, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Second query to update lastSeen
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userId, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen",
			"registered_at", "updated_at"}).
			AddRow(1, "johndoe", "john@example.com", "hashedpassword",
				"https://default-image.jpg", true, false, now, now, now))

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs("johndoe", "john@example.com", "hashedpassword",
			"https://default-image.jpg", true, false, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.MarkOffline(userId)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestSearchUsers_Success() {
	// arrange
	currentUserID := "1"
	query := "john"
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "picture_url",
		"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"})

	// Add two matching users
	rows.AddRow(2, "johndoe", "johndoe@example.com", "hashedpassword",
		"https://default-image.jpg", true, false, now, now, now)
	rows.AddRow(3, "johnsmith", "johnsmith@example.com", "hashedpassword",
		"https://default-image.jpg", true, true, now, now, now)

	s.mock.ExpectQuery(`SELECT "users"."id","users"."username","users"."email","users"."password","users"."picture_url","users"."is_email_valid","users"."is_online","users"."last_seen","users"."registered_at","users"."updated_at" FROM "users" LEFT JOIN user_friends ON users.id = user_friends.friend_id AND user_friends.user_id = \$1 WHERE users.username LIKE \$2 AND user_friends.friend_id IS NULL AND users.id != \$3 LIMIT \$4`).
		WithArgs(currentUserID, "%"+query+"%", currentUserID, 10).
		WillReturnRows(rows)

	// act
	users, err := s.userService.SearchUsers(currentUserID, query)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), users)
	assert.Equal(s.T(), 2, len(*users))
	assert.Equal(s.T(), "johndoe", (*users)[0].Username)
	assert.Equal(s.T(), "johnsmith", (*users)[1].Username)
}

func (s *UserServiceTestSuite) TestSearchUsers_NoResults() {
	// arrange
	currentUserID := "1"
	query := "xyz"

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "picture_url",
		"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"})

	s.mock.ExpectQuery(`SELECT "users"."id","users"."username","users"."email","users"."password","users"."picture_url","users"."is_email_valid","users"."is_online","users"."last_seen","users"."registered_at","users"."updated_at" FROM "users" LEFT JOIN user_friends ON users.id = user_friends.friend_id AND user_friends.user_id = \$1 WHERE users.username LIKE \$2 AND user_friends.friend_id IS NULL AND users.id != \$3 LIMIT \$4`).
		WithArgs(currentUserID, "%"+query+"%", currentUserID, 10).
		WillReturnRows(rows)

	// act
	users, err := s.userService.SearchUsers(currentUserID, query)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), users)
	assert.Equal(s.T(), 0, len(*users))
}

func (s *UserServiceTestSuite) TestSearchUsers_DatabaseError() {
	// arrange
	currentUserID := "1"
	query := "john"

	s.mock.ExpectQuery(`SELECT "users"."id","users"."username","users"."email","users"."password","users"."picture_url","users"."is_email_valid","users"."is_online","users"."last_seen","users"."registered_at","users"."updated_at" FROM "users" LEFT JOIN user_friends ON users.id = user_friends.friend_id AND user_friends.user_id = \$1 WHERE users.username LIKE \$2 AND user_friends.friend_id IS NULL AND users.id != \$3 LIMIT \$4`).
		WithArgs(currentUserID, "%"+query+"%", currentUserID, 10).
		WillReturnError(errors.New("database error"))

	// act
	users, err := s.userService.SearchUsers(currentUserID, query)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), users)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *UserServiceTestSuite) TestAcceptFriendRequest_Success() {
	// arrange
	requestID := uint(1)
	senderID := uint(2)
	receiverID := uint(3)
	now := time.Now()

	// Fetch the friend request
	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE "friend_requests"."id" = \$1 ORDER BY "friend_requests"."id" LIMIT \$2`).
		WithArgs(requestID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "status", "sent_at"}).
			AddRow(requestID, senderID, receiverID, "pending", now))

	// Update the request status
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "friend_requests" SET "sender_id"=\$1,"receiver_id"=\$2,"status"=\$3,"sent_at"=\$4,"responded_at"=\$5 WHERE "id" = \$6`).
		WithArgs(senderID, receiverID, "accepted", now, sqlmock.AnyArg(), requestID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Fetch sender
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(senderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(senderID, "sender", "sender@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Fetch receiver
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(receiverID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(receiverID, "receiver", "receiver@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// For first association operation (add receiver to sender's friends)
	s.mock.ExpectBegin()

	// GORM will update user record first
	s.mock.ExpectExec(`UPDATE "users" SET "updated_at"=\$1 WHERE "id" = \$2`).
		WithArgs(sqlmock.AnyArg(), senderID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// GORM may try to insert/check the user record
	s.mock.ExpectQuery(`INSERT INTO "users".* ON CONFLICT.*RETURNING "id"`).
		WithArgs("receiver", "receiver@example.com", "hashedpw", "https://default-image.jpg",
			true, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), receiverID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(receiverID))

	s.mock.ExpectExec(`INSERT INTO "user_friends".*`).
		WithArgs(senderID, receiverID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()

	// For second association operation (add sender to receiver's friends)
	s.mock.ExpectBegin()

	s.mock.ExpectExec(`UPDATE "users" SET "updated_at"=\$1 WHERE "id" = \$2`).
		WithArgs(sqlmock.AnyArg(), receiverID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectQuery(`INSERT INTO "users".*ON CONFLICT.*RETURNING "id"`).
		WithArgs("sender", "sender@example.com", "hashedpw", "https://default-image.jpg",
			true, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), senderID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(senderID))

	s.mock.ExpectExec(`INSERT INTO "user_friends" .*`).
		WithArgs(receiverID, senderID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()

	// act
	err := s.userService.AcceptFriendRequest(requestID)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestAcceptFriendRequest_NotFound() {
	// arrange
	requestID := uint(99)

	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE "friend_requests"."id" = \$1 ORDER BY "friend_requests"."id" LIMIT \$2`).
		WithArgs(requestID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	err := s.userService.AcceptFriendRequest(requestID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}

func (s *UserServiceTestSuite) TestAcceptFriendRequest_AlreadyProcessed() {
	// arrange
	requestID := uint(1)
	senderID := uint(2)
	receiverID := uint(3)
	now := time.Now()

	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE "friend_requests"."id" = \$1 ORDER BY "friend_requests"."id" LIMIT \$2`).
		WithArgs(requestID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "status", "sent_at", "responded_at"}).
			AddRow(requestID, senderID, receiverID, "accepted", now, now))

	// act
	err := s.userService.AcceptFriendRequest(requestID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "friend request already processed", err.Error())
}

func (s *UserServiceTestSuite) TestRejectFriendRequest_Success() {
	// arrange
	requestID := uint(1)
	senderID := uint(2)
	receiverID := uint(3)
	now := time.Now()

	// Fetch the friend request
	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE "friend_requests"."id" = \$1 ORDER BY "friend_requests"."id" LIMIT \$2`).
		WithArgs(requestID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "status", "sent_at"}).
			AddRow(requestID, senderID, receiverID, "pending", now))

	// Update the request status
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "friend_requests" SET "sender_id"=\$1,"receiver_id"=\$2,"status"=\$3,"sent_at"=\$4,"responded_at"=\$5 WHERE "id" = \$6`).
		WithArgs(senderID, receiverID, "rejected", now, sqlmock.AnyArg(), requestID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.userService.RejectFriendRequest(requestID)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestRejectFriendRequest_NotFound() {
	// arrange
	requestID := uint(99)

	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE "friend_requests"."id" = \$1 ORDER BY "friend_requests"."id" LIMIT \$2`).
		WithArgs(requestID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	err := s.userService.RejectFriendRequest(requestID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}

func (s *UserServiceTestSuite) TestRejectFriendRequest_AlreadyProcessed() {
	// arrange
	requestID := uint(1)
	senderID := uint(2)
	receiverID := uint(3)
	now := time.Now()

	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE "friend_requests"."id" = \$1 ORDER BY "friend_requests"."id" LIMIT \$2`).
		WithArgs(requestID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "status", "sent_at", "responded_at"}).
			AddRow(requestID, senderID, receiverID, "rejected", now, now))

	// act
	err := s.userService.RejectFriendRequest(requestID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "friend request already processed", err.Error())
}

func (s *UserServiceTestSuite) TestRemoveFriend_Success() {
	// arrange
	userID := uint(1)
	friendID := uint(2)
	now := time.Now()

	// Fetch the user
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(userID, "user", "user@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Fetch the friend
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(friendID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(friendID, "friend", "friend@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	s.mock.ExpectBegin()

	// Remove friend from user's friends
	s.mock.ExpectExec(`DELETE FROM "user_friends" WHERE "user_friends"."user_id" = \$1 AND "user_friends"."friend_id" = \$2`).
		WithArgs(userID, friendID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()

	s.mock.ExpectBegin()

	// Remove user from friend's friends
	s.mock.ExpectExec(`DELETE FROM "user_friends" WHERE "user_friends"."user_id" = \$1 AND "user_friends"."friend_id" = \$2`).
		WithArgs(friendID, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()

	// act
	err := s.userService.RemoveFriend(userID, friendID)

	// assert
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestRemoveFriend_UserNotFound() {
	// arrange
	userID := uint(99)
	friendID := uint(2)

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	err := s.userService.RemoveFriend(userID, friendID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}

func (s *UserServiceTestSuite) TestRemoveFriend_FriendNotFound() {
	// arrange
	userID := uint(1)
	friendID := uint(99)
	now := time.Now()

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(userID, "user", "user@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(friendID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	err := s.userService.RemoveFriend(userID, friendID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}

func (s *UserServiceTestSuite) TestGetFriends_Success() {
	// arrange
	userID := "1"
	now := time.Now()

	// Fetch the user
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(1, "user", "user@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Get friends from association
	s.mock.ExpectQuery(`SELECT "users"."id","users"."username","users"."email","users"."password",` +
		`"users"."picture_url","users"."is_email_valid","users"."is_online","users"."last_seen",` +
		`"users"."registered_at","users"."updated_at" FROM "users" JOIN "user_friends" ON ` +
		`"user_friends"."friend_id" = "users"."id" AND "user_friends"."user_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(2, "friend1", "friend1@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now).
			AddRow(3, "friend2", "friend2@example.com", "hashedpw",
				"https://default-image.jpg", true, true, now, now, now))

	// act
	friends, err := s.userService.GetFriends(userID)

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(friends))
	assert.Equal(s.T(), "friend1", friends[0].Username)
	assert.Equal(s.T(), "friend2", friends[1].Username)
}

func (s *UserServiceTestSuite) TestGetFriends_UserNotFound() {
	// arrange
	userID := "99"

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	friends, err := s.userService.GetFriends(userID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
	assert.Nil(s.T(), friends)
}

func (s *UserServiceTestSuite) TestGetFriendRequestsByStatus_Success() {
	// arrange
	userID := uint(1)
	status := "pending"
	now := time.Now()

	// Fetch friend requests
	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE receiver_id = \$1 AND status = \$2`).
		WithArgs(userID, status).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "status", "sent_at"}).
			AddRow(1, 2, userID, status, now).
			AddRow(2, 3, userID, status, now))

	// Preload Receivers
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(userID, "receiver", "receiver@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Preload Senders
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2\)`).
		WithArgs(2, 3).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(2, "sender1", "sender1@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now).
			AddRow(3, "sender2", "sender2@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// act
	requests, err := s.userService.GetFriendRequestsByStatus(userID, status)

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(*requests))
	assert.Equal(s.T(), uint(2), (*requests)[0].SenderID)
	assert.Equal(s.T(), uint(3), (*requests)[1].SenderID)
	assert.Equal(s.T(), status, (*requests)[0].Status)
}

func (s *UserServiceTestSuite) TestGetFriendRequestsByStatus_InvalidStatus() {
	// arrange
	userID := uint(1)
	status := "invalid"

	// act
	requests, err := s.userService.GetFriendRequestsByStatus(userID, status)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "invalid status", err.Error())
	assert.Nil(s.T(), requests)
}

func (s *UserServiceTestSuite) TestGetFriendRequestsByStatus_DatabaseError() {
	// arrange
	userID := uint(1)
	status := "pending"

	s.mock.ExpectQuery(`SELECT \* FROM "friend_requests" WHERE receiver_id = \$1 AND status = \$2`).
		WithArgs(userID, status).
		WillReturnError(errors.New("database error"))

	// act
	requests, err := s.userService.GetFriendRequestsByStatus(userID, status)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), requests)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *UserServiceTestSuite) TestCountPendingFriendRequests_Success() {
	// arrange
	userID := uint(1)
	expectedCount := int64(2)

	s.mock.ExpectQuery(`SELECT count\(\*\) FROM "friend_requests" WHERE receiver_id = \$1 AND status = \$2`).
		WithArgs(userID, "pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

	// act
	count, err := s.userService.CountPendingFriendRequests(userID)

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedCount, count)
}

func (s *UserServiceTestSuite) TestCountPendingFriendRequests_DatabaseError() {
	// arrange
	userID := uint(1)

	s.mock.ExpectQuery(`SELECT count\(\*\) FROM "friend_requests" WHERE receiver_id = \$1 AND status = \$2`).
		WithArgs(userID, "pending").
		WillReturnError(errors.New("database error"))

	// act
	count, err := s.userService.CountPendingFriendRequests(userID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), int64(0), count)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *UserServiceTestSuite) TestGetNumFriends_Success() {
	// arrange
	userID := "1"
	expectedCount := int64(3)
	now := time.Now()

	// Fetch the user
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(1, "user", "user@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Count association query
	s.mock.ExpectQuery(`SELECT count\(\*\) FROM "users" JOIN "user_friends" ON "user_friends"."friend_id" = "users"."id" AND "user_friends"."user_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

	// act
	count, err := s.userService.GetNumFriends(userID)

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedCount, count)
}

func (s *UserServiceTestSuite) TestGetNumFriends_UserNotFound() {
	// arrange
	userID := "999"

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	count, err := s.userService.GetNumFriends(userID)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), int64(0), count)
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}

func (s *UserServiceTestSuite) TestIsFriend_True() {
	// arrange
	userID := uint(1)
	friendID := uint(2)
	now := time.Now()

	// Fetch the user
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(userID, "user", "user@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Fetch the friend
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(friendID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(friendID, "friend", "friend@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Check association
	s.mock.ExpectQuery(`SELECT "users"."id","users"."username","users"."email","users"."password","users"."picture_url","users"."is_email_valid","users"."is_online","users"."last_seen","users"."registered_at","users"."updated_at" FROM "users" JOIN "user_friends" ON "user_friends"."friend_id" = "users"."id" AND "user_friends"."user_id" = \$1 WHERE id = \$2 AND "users"."id" = \$3`).
		WithArgs(userID, friendID, friendID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(friendID, "friend", "friend@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// act
	result := s.userService.IsFriend(userID, friendID)

	// assert
	assert.True(s.T(), result)
}

func (s *UserServiceTestSuite) TestIsFriend_False() {
	// arrange
	userID := uint(1)
	friendID := uint(2)
	now := time.Now()

	// Fetch the user
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(userID, "user", "user@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Fetch the friend
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(friendID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(friendID, "friend", "friend@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	// Check association (returns no rows)
	s.mock.ExpectQuery(`SELECT "users"."id","users"."username","users"."email","users"."password","users"."picture_url","users"."is_email_valid","users"."is_online","users"."last_seen","users"."registered_at","users"."updated_at" FROM "users" JOIN "user_friends" ON "user_friends"."friend_id" = "users"."id" AND "user_friends"."user_id" = \$1 WHERE id = \$2 AND "users"."id" = \$3`).
		WithArgs(userID, friendID, friendID).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	result := s.userService.IsFriend(userID, friendID)

	// assert
	assert.False(s.T(), result)
}

func (s *UserServiceTestSuite) TestIsFriend_UserNotFound() {
	// arrange
	userID := uint(999)
	friendID := uint(2)

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	result := s.userService.IsFriend(userID, friendID)

	// assert
	assert.False(s.T(), result)
}

func (s *UserServiceTestSuite) TestIsFriend_FriendNotFound() {
	// arrange
	userID := uint(1)
	friendID := uint(999)
	now := time.Now()

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "picture_url",
			"is_email_valid", "is_online", "last_seen", "registered_at", "updated_at"}).
			AddRow(userID, "user", "user@example.com", "hashedpw",
				"https://default-image.jpg", true, false, now, now, now))

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(friendID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	result := s.userService.IsFriend(userID, friendID)

	// assert
	assert.False(s.T(), result)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
