package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type SubscriptionRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	db           *gorm.DB
	logger       *logrus.Logger
	dependencies *tests.TestDependencies
	repo         SubscriptionRepository

	testUser *model.User
}

func (suite *SubscriptionRepositoryTestSuite) SetupSuite() {
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

	suite.repo = NewSubscriptionRepository(suite.db)
}

func (suite *SubscriptionRepositoryTestSuite) TearDownSuite() {
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *SubscriptionRepositoryTestSuite) SetupTest() {
	suite.testUser = &model.User{
		Username: "sub_user",
		Email:    "sub@example.com",
		Password: "password",
	}
	err := suite.db.Create(suite.testUser).Error
	assert.NoError(suite.T(), err)
}

func (suite *SubscriptionRepositoryTestSuite) TearDownTest() {
	suite.db.Exec("TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func TestSubscriptionRepositorySuite(t *testing.T) {
	suite.Run(t, new(SubscriptionRepositoryTestSuite))
}

func (suite *SubscriptionRepositoryTestSuite) TestCreate_Success() {
	sub := &model.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: "https://push.example.com/abc",
		P256dh:   "p256dh-key",
		Auth:     "auth-key",
	}
	created, err := suite.repo.Create(sub)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), sub.Endpoint, created.Endpoint)
	assert.NotZero(suite.T(), created.ID)
}

func (suite *SubscriptionRepositoryTestSuite) TestFindByID_Success() {
	sub := &model.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: "https://find.me/123",
		P256dh:   "pkey",
		Auth:     "akey",
	}
	created, err := suite.repo.Create(sub)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(created.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), created.ID, found.ID)
	assert.Equal(suite.T(), sub.Endpoint, found.Endpoint)
}

func (suite *SubscriptionRepositoryTestSuite) TestFindByUserID_Success() {
	sub := &model.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: "https://find.user",
		P256dh:   "k1",
		Auth:     "k2",
	}
	_, err := suite.repo.Create(sub)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByUserID(fmt.Sprintf("%d", suite.testUser.ID))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *found, 1)
	assert.Equal(suite.T(), sub.Endpoint, (*found)[0].Endpoint)
}

func (suite *SubscriptionRepositoryTestSuite) TestFindByEndpoint_Success() {
	endpoint := "https://endpoint.example.com/xyz"
	sub := &model.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: endpoint,
		P256dh:   "pkey",
		Auth:     "akey",
	}
	_, err := suite.repo.Create(sub)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByEndpoint(endpoint)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), endpoint, found.Endpoint)
}

func (suite *SubscriptionRepositoryTestSuite) TestDelete_Success() {
	sub := &model.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: "https://delete.me",
		P256dh:   "k3",
		Auth:     "k4",
	}
	created, err := suite.repo.Create(sub)
	assert.NoError(suite.T(), err)

	err = suite.repo.Delete(created.ID)
	assert.NoError(suite.T(), err)

	// should not find it again
	_, err = suite.repo.FindByEndpoint("https://delete.me")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}
