package test_services

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/RowenTey/JustJio/services"
)

type NotificationServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	notificationService *services.NotificationService
}

func (s *NotificationServiceTestSuite) SetupTest() {
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

	s.notificationService = &services.NotificationService{DB: s.DB}
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
	s.mock.ExpectQuery(`INSERT INTO "notifications" \("user_id","title","content","created_at","is_read","updated_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6\) RETURNING "id"`).
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
	s.mock.ExpectQuery(`INSERT INTO "notifications" \("user_id","title","content","created_at","is_read","updated_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6\) RETURNING "id"`).
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

func TestNotificationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceTestSuite))
}
