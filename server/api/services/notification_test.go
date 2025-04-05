package services

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type NotificationServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	notificationService *NotificationService

	userId uint
	title  string
}

func TestNotificationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceTestSuite))
}

func (s *NotificationServiceTestSuite) SetupTest() {
	var err error
	s.DB, s.mock, err = tests.SetupTestDB()
	assert.NoError(s.T(), err)

	s.notificationService = NewNotificationService(s.DB)

	s.userId = uint(1)
	s.title = "Test Title"
}

func (s *NotificationServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *NotificationServiceTestSuite) TestCreateNotification_Success() {
	// arrange
	content := "Test Content"
	now := time.Now()

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "notifications"`).
		WithArgs(
			s.userId,
			s.title,
			content,
			sqlmock.AnyArg(), // createdAt
			false,            // isRead
			sqlmock.AnyArg(), // updatedAt
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "content", "created_at", "is_read", "updated_at"}).
			AddRow(1, s.userId, s.title, content, now, false, now))
	s.mock.ExpectCommit()

	// act
	result, err := s.notificationService.CreateNotification(s.userId, s.title, content)

	// assert
	tests.AssertNoErrAndNotNil(s.T(), err, result)
	assert.Equal(s.T(), s.userId, result.UserID)
	assert.Equal(s.T(), s.title, result.Title)
	assert.Equal(s.T(), content, result.Content)
	assert.Equal(s.T(), false, result.IsRead)
}

func (s *NotificationServiceTestSuite) TestCreateNotification_EmptyContent() {
	// arrange
	content := "" // Empty content

	// act
	result, err := s.notificationService.CreateNotification(s.userId, s.title, content)

	// assert
	tests.AssertErrAndNil(s.T(), err, result)
	assert.Equal(s.T(), "content cannot be empty", err.Error())
}

func (s *NotificationServiceTestSuite) TestCreateNotification_DatabaseError() {
	// arrange
	content := "Test Content"

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "notifications"`).
		WithArgs(
			s.userId,
			s.title,
			content,
			sqlmock.AnyArg(), // createdAt
			false,            // isRead
			sqlmock.AnyArg(), // updatedAt
		).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	result, err := s.notificationService.CreateNotification(s.userId, s.title, content)

	// assert
	tests.AssertErrAndNil(s.T(), err, result)
	assert.Contains(s.T(), err.Error(), "database error")
}
