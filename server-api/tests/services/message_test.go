package test_services

import (
	"database/sql"
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

	"github.com/RowenTey/JustJio/model"
	"github.com/RowenTey/JustJio/services"
)

type MessageServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	messageService *services.MessageService
}

func (s *MessageServiceTestSuite) SetupTest() {
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

	s.messageService = &services.MessageService{DB: s.DB}
}

func (s *MessageServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *MessageServiceTestSuite) TestSaveMessage_Success() {
	// arrange
	room := &model.Room{
		ID:   "room-uuid-123",
		Name: "Test Room",
	}
	sender := &model.User{
		ID:       1,
		Username: "testuser",
	}
	content := "Hello, world!"

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "messages" \("room_id","sender_id","content","sent_at"\) VALUES \(\$1,\$2,\$3,\$4\) RETURNING "id"`).
		WithArgs(
			room.ID,          // RoomID (string UUID)
			sender.ID,        // SenderID
			content,          // Content
			sqlmock.AnyArg(), // SentAt
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // ID = 1
	s.mock.ExpectCommit()

	// act
	err := s.messageService.SaveMessage(room, sender, content)

	// assert
	assert.NoError(s.T(), err)
}

func (s *MessageServiceTestSuite) TestGetMessageById_Success() {
	// arrange
	messageID := "1"
	roomID := "room-uuid-123"
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "room_id", "sender_id", "content", "sent_at",
	}).AddRow(
		messageID, roomID, 1, "Hello, world!", now,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "messages" WHERE id = \$1 AND room_id = \$2 ORDER BY "messages"."id" LIMIT \$3`).
		WithArgs(messageID, roomID, 1).
		WillReturnRows(rows)

	// act
	message, err := s.messageService.GetMessageById(messageID, roomID)

	// assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), uint(1), message.ID)
	assert.Equal(s.T(), roomID, message.RoomID)
	assert.Equal(s.T(), uint(1), message.SenderID)
	assert.Equal(s.T(), "Hello, world!", message.Content)
	assert.Equal(s.T(), now.Unix(), message.SentAt.Unix())
}

func (s *MessageServiceTestSuite) TestDeleteMessage_Success() {
	// arrange
	messageID := "1"
	roomID := "room-uuid-123"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "messages" WHERE id = \$1 AND room_id = \$2`).
		WithArgs(messageID, roomID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.messageService.DeleteMessage(messageID, roomID)

	// assert
	assert.NoError(s.T(), err)
}

func (s *MessageServiceTestSuite) TestDeleteRoomMessages_Success() {
	// arrange
	roomID := "room-uuid-123"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "messages" WHERE room_id = \$1`).
		WithArgs(roomID).
		WillReturnResult(sqlmock.NewResult(0, 5)) // 5 messages deleted
	s.mock.ExpectCommit()

	// act
	err := s.messageService.DeleteRoomMessages(roomID)

	// assert
	assert.NoError(s.T(), err)
}

func TestMessageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MessageServiceTestSuite))
}
