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

type BillRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	db           *gorm.DB
	repo         BillRepository
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	testUser *model.User
	testRoom *model.Room
}

func (suite *BillRepositoryTestSuite) SetupSuite() {
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

	suite.repo = NewBillRepository(suite.db)
}

func (suite *BillRepositoryTestSuite) TearDownSuite() {
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *BillRepositoryTestSuite) SetupTest() {
	suite.testUser = &model.User{
		Username: "billuser",
		Email:    "bill@example.com",
		Password: "secret",
	}
	err := suite.db.Create(suite.testUser).Error
	assert.NoError(suite.T(), err)

	suite.testRoom = &model.Room{
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
	suite.Run(t, new(BillRepositoryTestSuite))
}

func (suite *BillRepositoryTestSuite) TestCreateAndFindByID_Success() {
	bill := model.Bill{
		Amount:  50.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err := suite.repo.Create(&bill)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(bill.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), bill.Amount, found.Amount)
}

func (suite *BillRepositoryTestSuite) TestFindByRoom_Success() {
	bill := model.Bill{
		Amount:  20.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err := suite.repo.Create(&bill)
	assert.NoError(suite.T(), err)

	bills, err := suite.repo.FindByRoom(suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *bills, 1)
}

func (suite *BillRepositoryTestSuite) TestDeleteByRoom_Success() {
	bill := model.Bill{
		Amount:  15.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err := suite.repo.Create(&bill)
	assert.NoError(suite.T(), err)

	err = suite.repo.DeleteByRoom(suite.testRoom.ID)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByRoom(suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *found, 0)
}

func (suite *BillRepositoryTestSuite) TestHasUnconsolidatedBills_Success() {
	// No bills yet
	status, err := suite.repo.GetRoomBillConsolidationStatus(suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), NO_BILLS, status)

	// Add an unconsolidated bill
	bill := model.Bill{
		Amount:  30.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err = suite.repo.Create(&bill)
	assert.NoError(suite.T(), err)

	status, err = suite.repo.GetRoomBillConsolidationStatus(suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), UNCONSOLIDATED, status)
}

func (suite *BillRepositoryTestSuite) TestConsolidateBills_Success() {
	bill := model.Bill{
		Amount:  60.0,
		RoomID:  suite.testRoom.ID,
		OwnerID: suite.testUser.ID,
	}
	err := suite.repo.Create(&bill)
	assert.NoError(suite.T(), err)

	consolidation, err := suite.repo.ConsolidateBills(suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), consolidation.ID)

	updated, err := suite.repo.FindByID(bill.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), consolidation.ID, updated.ConsolidationID)

	// Confirm status
	status, err := suite.repo.GetRoomBillConsolidationStatus(suite.testRoom.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), CONSOLIDATED, status)
}

func (suite *BillRepositoryTestSuite) TestFindByConsolidation_Success() {
	consolidation := model.Consolidation{}
	err := suite.db.Create(&consolidation).Error
	assert.NoError(suite.T(), err)

	bill := model.Bill{
		Amount:          100,
		RoomID:          suite.testRoom.ID,
		OwnerID:         suite.testUser.ID,
		ConsolidationID: consolidation.ID,
	}
	err = suite.repo.Create(&bill)
	assert.NoError(suite.T(), err)

	bills, err := suite.repo.FindByConsolidation(consolidation.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *bills, 1)
}
