package repositories

import (
	"context"
	"testing"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/pkg/tests"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type BillRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	db           *gorm.DB
	repo         BillRepository
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	testUser *models.User
	testRoom *models.Room
}

func (suite *BillRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Setup DB Conn
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "bill_test", "file://../../migrations")
	assert.NoError(suite.T(), err)

	suite.repo = NewBillRepository(suite.db)
}

func (suite *BillRepositoryTestSuite) TearDownSuite() {
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *BillRepositoryTestSuite) SetupTest() {
	suite.testUser = &models.User{
		Username: "billuser",
		Email:    "bill@example.com",
		Password: "secret",
	}
	err := suite.db.Create(suite.testUser).Error
	assert.NoError(suite.T(), err)

	suite.testRoom = &models.Room{
		Name:   "BillsRoom",
		HostID: suite.testUser.ID,
	}
	err = suite.db.Create(suite.testRoom).Error
	assert.NoError(suite.T(), err)
}

func (suite *BillRepositoryTestSuite) TearDownTest() {
	suite.db.Exec("TRUNCATE TABLE bills RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE consolidations RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms RESTART IDENTITY CASCADE")
}

func TestBillRepositorySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(BillRepositoryTestSuite))
}

func (suite *BillRepositoryTestSuite) TestCreateAndFindByID_Success() {
	bill := models.Bill{
		Amount:  50.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err := suite.repo.Create(suite.ctx, &bill)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(suite.ctx, bill.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), bill.Amount, found.Amount)
}

func (suite *BillRepositoryTestSuite) TestFindByRoom_Success() {
	bill := models.Bill{
		Amount:  20.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err := suite.repo.Create(suite.ctx, &bill)
	assert.NoError(suite.T(), err)

	bills, err := suite.repo.FindByRoom(suite.ctx, suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), bills, 1)
}

func (suite *BillRepositoryTestSuite) TestDeleteByRoom_Success() {
	bill := models.Bill{
		Amount:  15.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err := suite.repo.Create(suite.ctx, &bill)
	assert.NoError(suite.T(), err)

	err = suite.repo.DeleteByRoom(suite.ctx, suite.testRoom.ID)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByRoom(suite.ctx, suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), found, 0)
}

func (suite *BillRepositoryTestSuite) TestConsolidateBills_Success() {
	bill := models.Bill{
		Amount:  60.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err := suite.repo.Create(suite.ctx, &bill)
	assert.NoError(suite.T(), err)

	consolidation, err := suite.repo.ConsolidateBills(suite.ctx, suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), consolidation.ID)

	updated, err := suite.repo.FindByID(suite.ctx, bill.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), consolidation.ID, updated.ConsolidationID)
}

func (suite *BillRepositoryTestSuite) TestFindByConsolidation_Success() {
	consolidation := models.Consolidation{}
	err := suite.db.Create(&consolidation).Error
	assert.NoError(suite.T(), err)

	bill := models.Bill{
		Amount:          100,
		RoomID:          suite.testRoom.ID,
		OwnerID:         suite.testUser.ID,
		ConsolidationID: consolidation.ID,
	}
	err = suite.repo.Create(suite.ctx, &bill)
	assert.NoError(suite.T(), err)

	bills, err := suite.repo.FindByConsolidation(suite.ctx, consolidation.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), bills, 1)
}
