package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MessageRepositoryTestSuite struct {
	suite.Suite
	db           *gorm.DB
	ctx          context.Context
	repo         MessageRepository
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	testRoom *model.Room
	testUser *model.User
}

func (suite *MessageRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)
	fmt.Println("Test DB Connection String:", pgConnStr)

	// Initialize database
	suite.db, err = database.InitTestDB(pgConnStr)
	assert.NoError(suite.T(), err)

	// Run migrations
	err = database.Migrate(suite.db)
	assert.NoError(suite.T(), err)

	suite.repo = NewMessageRepository(suite.db)
}

func (suite *MessageRepositoryTestSuite) TearDownSuite() {
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *MessageRepositoryTestSuite) SetupTest() {
	suite.testUser = &model.User{
		Username: "msg_user",
		Email:    "msg@example.com",
		Password: "securepass",
	}
	err := suite.db.Create(suite.testUser).Error
	assert.NoError(suite.T(), err)

	suite.testRoom = &model.Room{
		Name:   "Test Room",
		HostID: suite.testUser.ID,
	}
	err = suite.db.Create(suite.testRoom).Error
	assert.NoError(suite.T(), err)
}

func (suite *MessageRepositoryTestSuite) TearDownTest() {
	suite.db.Exec("TRUNCATE TABLE messages RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func TestMessageRepositorySuite(t *testing.T) {
	suite.Run(t, new(MessageRepositoryTestSuite))
}

func (suite *MessageRepositoryTestSuite) TestCreateAndFindByID_Success() {
	message := model.Message{
		Content:  "Hello, world!",
		RoomID:   suite.testRoom.ID,
		SenderID: suite.testUser.ID,
		SentAt:   time.Now(),
	}

	err := suite.repo.Create(&message)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(fmt.Sprintf("%d", message.ID), suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), message.Content, found.Content)
	assert.Equal(suite.T(), message.RoomID, found.RoomID)
}

func (suite *MessageRepositoryTestSuite) TestDelete_Success() {
	message := model.Message{
		Content:  "To be deleted",
		RoomID:   suite.testRoom.ID,
		SenderID: suite.testUser.ID,
		SentAt:   time.Now(),
	}
	err := suite.repo.Create(&message)
	assert.NoError(suite.T(), err)

	err = suite.repo.Delete(fmt.Sprintf("%d", message.ID), suite.testRoom.ID)
	assert.NoError(suite.T(), err)

	_, err = suite.repo.FindByID(fmt.Sprintf("%d", message.ID), suite.testRoom.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *MessageRepositoryTestSuite) TestDeleteByRoom_Success() {
	for i := 0; i < 3; i++ {
		msg := model.Message{
			Content:  fmt.Sprintf("msg-%d", i),
			RoomID:   suite.testRoom.ID,
			SenderID: suite.testUser.ID,
			SentAt:   time.Now(),
		}
		err := suite.repo.Create(&msg)
		assert.NoError(suite.T(), err)
	}

	err := suite.repo.DeleteByRoom(suite.testRoom.ID)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountByRoom(suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *MessageRepositoryTestSuite) TestCountByRoom_Success() {
	msg := model.Message{
		Content:  "count me",
		RoomID:   suite.testRoom.ID,
		SenderID: suite.testUser.ID,
		SentAt:   time.Now(),
	}
	err := suite.repo.Create(&msg)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountByRoom(suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *MessageRepositoryTestSuite) TestFindByRoom_WithPaginationAndOrder() {
	// Create 5 messages with ascending timestamps
	for i := 0; i < 5; i++ {
		msg := model.Message{
			Content:  fmt.Sprintf("msg-%d", i),
			RoomID:   suite.testRoom.ID,
			SenderID: suite.testUser.ID,
			SentAt:   time.Now().Add(time.Duration(i) * time.Minute),
		}
		err := suite.repo.Create(&msg)
		assert.NoError(suite.T(), err)
	}

	// Get first 3 messages in descending order
	msgs, err := suite.repo.FindByRoom(suite.testRoom.ID, 1, 3, false)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *msgs, 3)
	assert.True(suite.T(), (*msgs)[0].SentAt.After((*msgs)[1].SentAt))

	// Get in ascending order
	msgsAsc, err := suite.repo.FindByRoom(suite.testRoom.ID, 1, 3, true)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *msgsAsc, 3)
	assert.True(suite.T(), (*msgsAsc)[0].SentAt.Before((*msgsAsc)[1].SentAt))
}
