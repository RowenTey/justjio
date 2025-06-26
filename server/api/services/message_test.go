package services

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MessageServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	messageService *MessageService

	roomId string
}

// func TestMessageServiceTestSuite(t *testing.T) {
// 	suite.Run(t, new(MessageServiceTestSuite))
// }

// func (s *MessageServiceTestSuite) SetupTest() {
// 	var err error
// 	s.DB, s.mock, err = tests.SetupTestDB()
// 	assert.NoError(s.T(), err)

// 	s.messageService = NewMessageService(s.DB)

// 	s.roomId = "room-123"
// }

// func (s *MessageServiceTestSuite) AfterTest(_, _ string) {
// 	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
// }

// func (s *MessageServiceTestSuite) TestSaveMessage_Success() {
// 	// arrange
// 	room := tests.CreateTestRoom(s.roomId, "Test Room", 1)
// 	sender := tests.CreateTestUser(1, "testuser", "user@test.com")
// 	content := "Hello, world!"

// 	s.mock.ExpectBegin()
// 	s.mock.ExpectQuery(`INSERT INTO "messages"`).
// 		WithArgs(
// 			room.ID,
// 			sender.ID,
// 			content,
// 			sqlmock.AnyArg(), // SentAt
// 		).
// 		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // ID = 1
// 	s.mock.ExpectCommit()

// 	// act
// 	err := s.messageService.SaveMessage(room, sender, content)

// 	// assert
// 	assert.NoError(s.T(), err)
// }

// func (s *MessageServiceTestSuite) TestGetMessageById_Success() {
// 	// arrange
// 	messageID := "1"
// 	now := time.Now()

// 	rows := sqlmock.NewRows([]string{
// 		"id", "room_id", "sender_id", "content", "sent_at",
// 	}).AddRow(
// 		messageID, s.roomId, 1, "Hello, world!", now,
// 	)

// 	s.mock.ExpectQuery(`SELECT \* FROM "messages" WHERE id = \$1 AND room_id = \$2 ORDER BY "messages"."id" LIMIT \$3`).
// 		WithArgs(messageID, s.roomId, 1).
// 		WillReturnRows(rows)

// 	// act
// 	message, err := s.messageService.GetMessageById(messageID, s.roomId)

// 	// assert
// 	assert.NoError(s.T(), err)
// 	assert.Equal(s.T(), uint(1), message.ID)
// 	assert.Equal(s.T(), s.roomId, message.RoomID)
// 	assert.Equal(s.T(), uint(1), message.SenderID)
// 	assert.Equal(s.T(), "Hello, world!", message.Content)
// 	assert.Equal(s.T(), now.Unix(), message.SentAt.Unix())
// }

// func (s *MessageServiceTestSuite) TestDeleteMessage_Success() {
// 	// arrange
// 	messageID := "1"

// 	s.mock.ExpectBegin()
// 	s.mock.ExpectExec(`DELETE FROM "messages" WHERE id = \$1 AND room_id = \$2`).
// 		WithArgs(messageID, s.roomId).
// 		WillReturnResult(sqlmock.NewResult(1, 1))
// 	s.mock.ExpectCommit()

// 	// act
// 	err := s.messageService.DeleteMessage(messageID, s.roomId)

// 	// assert
// 	assert.NoError(s.T(), err)
// }

// func (s *MessageServiceTestSuite) TestDeleteRoomMessages_Success() {
// 	// arrange
// 	s.mock.ExpectBegin()
// 	s.mock.ExpectExec(`DELETE FROM "messages" WHERE room_id = \$1`).
// 		WithArgs(s.roomId).
// 		WillReturnResult(sqlmock.NewResult(0, 5)) // 5 messages deleted
// 	s.mock.ExpectCommit()

// 	// act
// 	err := s.messageService.DeleteRoomMessages(s.roomId)

// 	// assert
// 	assert.NoError(s.T(), err)
// }
