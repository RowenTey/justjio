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

type SubscriptionRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	db           *gorm.DB
	logger       *logrus.Logger
	dependencies *tests.TestDependencies
	repo         SubscriptionRepository

	testUser *models.User
}

func (suite *SubscriptionRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Setup DB Conn
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "sub_test", "file://../../migrations")
	assert.NoError(suite.T(), err)

	suite.repo = NewSubscriptionRepository(suite.db)
}

func (suite *SubscriptionRepositoryTestSuite) TearDownSuite() {
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *SubscriptionRepositoryTestSuite) SetupTest() {
	suite.testUser = &models.User{
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
	t.Parallel()
	suite.Run(t, new(SubscriptionRepositoryTestSuite))
}

func (suite *SubscriptionRepositoryTestSuite) TestCreate_Success() {
	sub := &models.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: "https://push.example.com/abc",
		P256dh:   "p256dh-key",
		Auth:     "auth-key",
	}
	created, err := suite.repo.Create(suite.ctx, sub)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), sub.Endpoint, created.Endpoint)
	assert.NotZero(suite.T(), created.ID)
}

func (suite *SubscriptionRepositoryTestSuite) TestFindByID_Success() {
	sub := &models.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: "https://find.me/123",
		P256dh:   "pkey",
		Auth:     "akey",
	}
	created, err := suite.repo.Create(suite.ctx, sub)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(suite.ctx, created.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), created.ID, found.ID)
	assert.Equal(suite.T(), sub.Endpoint, found.Endpoint)
}

func (suite *SubscriptionRepositoryTestSuite) TestFindByUserID_Success() {
	sub := &models.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: "https://find.user",
		P256dh:   "k1",
		Auth:     "k2",
	}
	_, err := suite.repo.Create(suite.ctx, sub)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByUserID(suite.ctx, fmt.Sprintf("%d", suite.testUser.ID))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), found, 1)
	assert.Equal(suite.T(), sub.Endpoint, found[0].Endpoint)
}

func (suite *SubscriptionRepositoryTestSuite) TestFindByEndpoint_Success() {
	endpoint := "https://endpoint.example.com/xyz"
	sub := &models.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: endpoint,
		P256dh:   "pkey",
		Auth:     "akey",
	}
	_, err := suite.repo.Create(suite.ctx, sub)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByEndpoint(suite.ctx, endpoint)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), endpoint, found.Endpoint)
}

func (suite *SubscriptionRepositoryTestSuite) TestDelete_Success() {
	sub := &models.Subscription{
		UserID:   suite.testUser.ID,
		Endpoint: "https://delete.me",
		P256dh:   "k3",
		Auth:     "k4",
	}
	created, err := suite.repo.Create(suite.ctx, sub)
	assert.NoError(suite.T(), err)

	err = suite.repo.Delete(suite.ctx, created.ID)
	assert.NoError(suite.T(), err)

	// should not find it again
	_, err = suite.repo.FindByEndpoint(suite.ctx, "https://delete.me")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}
