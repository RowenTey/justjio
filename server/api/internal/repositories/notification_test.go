package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/pkg/tests"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type NotificationRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	db           *gorm.DB
	repo         NotificationRepository
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	testUser *models.User
}

func (suite *NotificationRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Setup DB Conn
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "noti_test", "file://../../migrations")
	assert.NoError(suite.T(), err)

	suite.repo = NewNotificationRepository(suite.db)
}

func (suite *NotificationRepositoryTestSuite) TearDownSuite() {
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *NotificationRepositoryTestSuite) SetupTest() {
	suite.testUser = &models.User{
		Username: "notify_user",
		Email:    "notify@example.com",
		Password: "securehash",
	}
	err := suite.db.Create(suite.testUser).Error
	assert.NoError(suite.T(), err)
}

func (suite *NotificationRepositoryTestSuite) TearDownTest() {
	suite.db.Exec("TRUNCATE TABLE notifications RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func TestNotificationRepositorySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(NotificationRepositoryTestSuite))
}

func (suite *NotificationRepositoryTestSuite) TestCreateNotification_Success() {
	notification := &models.Notification{
		UserID:  suite.testUser.ID,
		Title:   "Welcome",
		Content: "You've been notified!",
	}
	created, err := suite.repo.Create(suite.ctx, notification)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), created.ID)
	assert.Equal(suite.T(), notification.Title, created.Title)
}

func (suite *NotificationRepositoryTestSuite) TestFindByIDAndUser_Success() {
	notification := &models.Notification{
		UserID:  suite.testUser.ID,
		Title:   "FindMe",
		Content: "This is searchable",
	}
	created, err := suite.repo.Create(suite.ctx, notification)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(suite.ctx, created.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), created.ID, found.ID)
}

func (suite *NotificationRepositoryTestSuite) TestFindByIDAndUser_NotFound() {
	_, err := suite.repo.FindByID(suite.ctx, 999)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *NotificationRepositoryTestSuite) TestFindByUser_Success() {
	notifications := []models.Notification{
		{UserID: suite.testUser.ID, Title: "A", Content: "A"},
		{UserID: suite.testUser.ID, Title: "B", Content: "B"},
	}
	for _, n := range notifications {
		err := suite.db.Create(&n).Error
		assert.NoError(suite.T(), err)
	}

	found, err := suite.repo.FindByUser(suite.ctx, fmt.Sprintf("%d", suite.testUser.ID))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), found, 2)
}

func (suite *NotificationRepositoryTestSuite) TestMarkAsRead_Success() {
	notification := &models.Notification{
		UserID:  suite.testUser.ID,
		Title:   "Unread",
		Content: "Still unread",
		IsRead:  false,
	}
	err := suite.db.Create(&notification).Error
	assert.NoError(suite.T(), err)

	err = suite.repo.MarkAsRead(suite.ctx, notification.ID)
	assert.NoError(suite.T(), err)

	var updated models.Notification
	err = suite.db.First(&updated, notification.ID).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updated.IsRead)
}
