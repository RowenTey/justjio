package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TransactionRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	db           *gorm.DB
	logger       *logrus.Logger
	dependencies *tests.TestDependencies
	repo         TransactionRepository

	userA         *model.User
	userB         *model.User
	bill          *model.Bill
	consolidation *model.Consolidation
}

func (suite *TransactionRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Setup DB Conn
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "tx_test")
	assert.NoError(suite.T(), err)

	suite.repo = NewTransactionRepository(suite.db)
}

func (suite *TransactionRepositoryTestSuite) TearDownSuite() {
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *TransactionRepositoryTestSuite) SetupTest() {
	suite.userA = &model.User{Username: "payer", Email: "payer@example.com", Password: "pw"}
	suite.userB = &model.User{Username: "payee", Email: "payee@example.com", Password: "pw"}

	err := suite.db.Create(suite.userA).Error
	assert.NoError(suite.T(), err)
	err = suite.db.Create(suite.userB).Error
	assert.NoError(suite.T(), err)

	suite.consolidation = &model.Consolidation{}
	err = suite.db.Create(suite.consolidation).Error
	assert.NoError(suite.T(), err)

	suite.bill = &model.Bill{
		Name:            "Test Bill",
		Amount:          100.0,
		Date:            time.Now(),
		RoomID:          "test-room-id",
		OwnerID:         suite.userA.ID,
		IncludeOwner:    true,
		Payers:          []model.User{*suite.userB},
		ConsolidationID: suite.consolidation.ID,
		Room:            model.Room{ID: "test-room-id", Name: "Test Room", HostID: suite.userA.ID},
	}
	err = suite.db.Create(suite.bill).Error
	assert.NoError(suite.T(), err)
}

func (suite *TransactionRepositoryTestSuite) TearDownTest() {
	suite.db.Exec("TRUNCATE TABLE transactions RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE consolidations RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE bills RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func TestTransactionRepositorySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TransactionRepositoryTestSuite))
}

func (suite *TransactionRepositoryTestSuite) TestCreateTransactions_Success() {
	txs := []model.Transaction{
		{
			ConsolidationID: suite.consolidation.ID,
			PayerID:         suite.userA.ID,
			PayeeID:         suite.userB.ID,
			Amount:          100.0,
			IsPaid:          false,
		},
		{
			ConsolidationID: suite.consolidation.ID,
			PayerID:         suite.userB.ID,
			PayeeID:         suite.userA.ID,
			Amount:          50.0,
			IsPaid:          true,
		},
	}

	err := suite.repo.Create(&txs)
	assert.NoError(suite.T(), err)

	var count int64
	err = suite.db.Model(&model.Transaction{}).Count(&count).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), count)
}

func (suite *TransactionRepositoryTestSuite) TestFindByUser_Unpaid() {
	tx := model.Transaction{
		ConsolidationID: suite.consolidation.ID,
		PayerID:         suite.userA.ID,
		PayeeID:         suite.userB.ID,
		Amount:          100.0,
		IsPaid:          false,
	}
	err := suite.repo.Create(&[]model.Transaction{tx})
	assert.NoError(suite.T(), err)

	txs, err := suite.repo.FindByUser(false, fmt.Sprint(suite.userA.ID))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *txs, 1)
	assert.Equal(suite.T(), false, (*txs)[0].IsPaid)
	assert.Equal(suite.T(), tx.Amount, (*txs)[0].Amount)
}

func (suite *TransactionRepositoryTestSuite) TestFindByUser_Paid() {
	tx := model.Transaction{
		ConsolidationID: suite.consolidation.ID,
		PayerID:         suite.userB.ID,
		PayeeID:         suite.userA.ID,
		Amount:          200.0,
		IsPaid:          true,
	}
	err := suite.repo.Create(&[]model.Transaction{tx})
	assert.NoError(suite.T(), err)

	txs, err := suite.repo.FindByUser(true, fmt.Sprint(suite.userA.ID))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *txs, 1)
	assert.Equal(suite.T(), true, (*txs)[0].IsPaid)
	assert.Equal(suite.T(), tx.Amount, (*txs)[0].Amount)
}

func (suite *TransactionRepositoryTestSuite) TestFindByID_Success() {
	tx := model.Transaction{
		// Supply ID directly for testing
		ID:              1,
		ConsolidationID: suite.consolidation.ID,
		PayerID:         suite.userA.ID,
		PayeeID:         suite.userB.ID,
		Amount:          75.0,
		IsPaid:          false,
	}
	err := suite.repo.Create(&[]model.Transaction{tx})
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(fmt.Sprintf("%d", tx.ID))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), tx.Amount, found.Amount)
	assert.Equal(suite.T(), tx.PayerID, found.PayerID)
}

func (suite *TransactionRepositoryTestSuite) TestUpdateTransaction_Success() {
	tx := model.Transaction{
		ConsolidationID: suite.consolidation.ID,
		PayerID:         suite.userA.ID,
		PayeeID:         suite.userB.ID,
		Amount:          10.0,
		IsPaid:          false,
	}
	err := suite.repo.Create(&[]model.Transaction{tx})
	assert.NoError(suite.T(), err)

	// Update IsPaid to true
	tx.IsPaid = true
	err = suite.repo.Update(&tx)
	assert.NoError(suite.T(), err)

	updated, err := suite.repo.FindByID(fmt.Sprintf("%d", tx.ID))
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updated.IsPaid)
}
