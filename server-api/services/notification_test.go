package services

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RowenTey/JustJio/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type NotificationServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	notificationService *NotificationService
}

func TestNotificationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceTestSuite))
}

func (s *NotificationServiceTestSuite) SetupTest() {
	var err error
	s.DB, s.mock, err = util.SetupTestDB()
	assert.NoError(s.T(), err)

	s.notificationService = NewNotificationService(s.DB)
}

func (s *NotificationServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *NotificationServiceTestSuite) TestCreateNotification_Success() {
	// arrange
	userId := uint(1)
	title := "Test Title"
	content := "Test Content"
	now := time.Now()

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "notifications"`).
		WithArgs(
			userId,
			title,
			content,
			sqlmock.AnyArg(), // createdAt
			false,            // isRead
			sqlmock.AnyArg(), // updatedAt
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "content", "created_at", "is_read", "updated_at"}).
			AddRow(1, userId, title, content, now, false, now))
	s.mock.ExpectCommit()

	// act
	result, err := s.notificationService.CreateNotification(userId, title, content)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), userId, result.UserID)
	assert.Equal(s.T(), title, result.Title)
	assert.Equal(s.T(), content, result.Content)
	assert.Equal(s.T(), false, result.IsRead)
}

func (s *NotificationServiceTestSuite) TestCreateNotification_EmptyContent() {
	// arrange
	userId := uint(1)
	title := "Test Title"
	content := "" // Empty content

	// act
	result, err := s.notificationService.CreateNotification(userId, title, content)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Equal(s.T(), "content cannot be empty", err.Error())
}

func (s *NotificationServiceTestSuite) TestCreateNotification_DatabaseError() {
	// arrange
	userId := uint(1)
	title := "Test Title"
	content := "Test Content"

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "notifications"`).
		WithArgs(
			userId,
			title,
			content,
			sqlmock.AnyArg(), // createdAt
			false,            // isRead
			sqlmock.AnyArg(), // updatedAt
		).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	result, err := s.notificationService.CreateNotification(userId, title, content)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}
