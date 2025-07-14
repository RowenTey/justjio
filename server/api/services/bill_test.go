package services

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/tests"
)

type BillServiceTestSuite struct {
	suite.Suite
	billService *BillService

	// DB mocks
	db      *gorm.DB
	sqlMock sqlmock.Sqlmock

	// Mock repositories
	mockBillRepo        *repository.MockBillRepository
	mockUserRepo        *repository.MockUserRepository
	mockRoomRepo        *repository.MockRoomRepository
	mockTransactionRepo *repository.MockTransactionRepository
	mockTransactionSvc  *MockTransactionService
}

func TestBillServiceSuite(t *testing.T) {
	suite.Run(t, new(BillServiceTestSuite))
}

func (s *BillServiceTestSuite) SetupTest() {
	var err error
	s.db, s.sqlMock, err = tests.SetupTestDB()
	require.NoError(s.T(), err)

	// Initialize mock repositories
	s.mockBillRepo = new(repository.MockBillRepository)
	s.mockUserRepo = new(repository.MockUserRepository)
	s.mockRoomRepo = new(repository.MockRoomRepository)
	s.mockTransactionRepo = new(repository.MockTransactionRepository)
	s.mockTransactionSvc = new(MockTransactionService)

	// Create billService with mock dependencies
	s.billService = NewBillService(
		s.db,
		s.mockBillRepo,
		s.mockUserRepo,
		s.mockRoomRepo,
		s.mockTransactionRepo,
		s.mockTransactionSvc,
		logrus.New(),
	)
}

func (s *BillServiceTestSuite) TestCreateBill_Success() {
	// Setup test data
	roomId := "room1"
	ownerId := "user1"
	payersId := []uint{2, 3}
	name := "Dinner"
	amount := float32(100.50)
	includeOwner := true

	room := &model.Room{ID: roomId}
	owner := &model.User{ID: 1, Username: "owner"}
	payers := []model.User{
		{ID: 2, Username: "payer1"},
		{ID: 3, Username: "payer2"},
	}

	// Mock expectations
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.NO_BILLS, nil)
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)
	s.mockUserRepo.On("FindByID", ownerId).Return(owner, nil)
	s.mockUserRepo.On("FindByIDs", &payersId).Return(&payers, nil)
	s.mockBillRepo.On("Create", mock.AnythingOfType("*model.Bill")).Run(func(args mock.Arguments) {
		bill := args.Get(0).(*model.Bill)
		bill.ID = 1 // Set ID for the created bill
	}).Return(nil)

	// Execute
	bill, err := s.billService.CreateBill(roomId, ownerId, &payersId, name, amount, includeOwner)

	// Assertions
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), bill)
	assert.Equal(s.T(), name, bill.Name)
	assert.Equal(s.T(), amount, bill.Amount)
	assert.Equal(s.T(), room.ID, bill.RoomID)
	assert.Equal(s.T(), owner.ID, bill.OwnerID)
	assert.Len(s.T(), bill.Payers, 2)

	// Verify mock calls
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *BillServiceTestSuite) TestCreateBill_AlreadyConsolidated() {
	roomId := "room1"
	ownerId := "user1"
	payersId := []uint{2, 3}

	// Mock expectations
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.CONSOLIDATED, nil)

	// Execute
	bill, err := s.billService.CreateBill(roomId, ownerId, &payersId, "Dinner", 100.50, true)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrAlreadyConsolidated, err)
	assert.Nil(s.T(), bill)

	// Verify mock calls
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertNotCalled(s.T(), "GetByID")
}

func (s *BillServiceTestSuite) TestCreateBill_EmptyPayers() {
	roomId := "room1"
	ownerId := "user1"
	emptyPayers := []uint{}

	// Mock expectations
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.NO_BILLS, nil)

	// Execute
	bill, err := s.billService.CreateBill(roomId, ownerId, &emptyPayers, "Dinner", 100.50, true)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrEmptyPayers, err)
	assert.Nil(s.T(), bill)

	// Verify mock calls
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertNotCalled(s.T(), "GetByID")
}

func (s *BillServiceTestSuite) TestGetBillById_Success() {
	billId := uint(1)
	expectedBill := &model.Bill{
		ID:     billId,
		Name:   "Test Bill",
		Amount: 50.0,
	}

	// Mock expectations
	s.mockBillRepo.On("FindByID", billId).Return(expectedBill, nil)

	// Execute
	bill, err := s.billService.GetBillById(billId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedBill, bill)
	s.mockBillRepo.AssertExpectations(s.T())
}

func (s *BillServiceTestSuite) TestGetBillsForRoom_Success() {
	roomId := "room1"
	expectedBills := []model.Bill{
		{ID: 1, Name: "Bill 1", RoomID: "1"},
		{ID: 2, Name: "Bill 2", RoomID: "1"},
	}

	// Mock expectations
	s.mockBillRepo.On("FindByRoom", roomId).Return(&expectedBills, nil)

	// Execute
	bills, err := s.billService.GetBillsForRoom(roomId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &expectedBills, bills)
	s.mockBillRepo.AssertExpectations(s.T())
}

func (s *BillServiceTestSuite) TestDeleteRoomBills_Success() {
	roomId := "room1"

	// Mock expectations
	s.mockBillRepo.On("DeleteByRoom", roomId).Return(nil)

	// Execute
	err := s.billService.DeleteRoomBills(roomId)

	// Assertions
	assert.NoError(s.T(), err)
	s.mockBillRepo.AssertExpectations(s.T())
}

func (s *BillServiceTestSuite) TestConsolidateBills_Success() {
	// Setup test data
	roomId := "room1"
	userId := "1"
	hostId := uint(1)

	room := &model.Room{ID: roomId, HostID: hostId}
	consolidation := &model.Consolidation{ID: 1}
	bills := []model.Bill{
		{ID: 1, Name: "Bill 1", Amount: 50.0, Owner: model.User{ID: 1}, Payers: []model.User{{ID: 2}}},
		{ID: 2, Name: "Bill 2", Amount: 30.0, Owner: model.User{ID: 2}, Payers: []model.User{{ID: 1}}},
	}
	transaction := &[]model.Transaction{{ID: 1, ConsolidationID: 1, Amount: 20.0, PayerID: 2, PayeeID: 1}}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockBillRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockBillRepo)
	s.mockTransactionRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockTransactionRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.NO_BILLS, nil)
	s.mockBillRepo.On("ConsolidateBills", roomId).Return(consolidation, nil)
	s.mockBillRepo.On("FindByConsolidation", consolidation.ID).Return(&bills, nil)
	s.mockTransactionSvc.On("GenerateTransactions", &bills, consolidation).Return(transaction, nil)
	s.mockTransactionRepo.On("Create", transaction).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.billService.ConsolidateBills(roomId, userId)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockTransactionRepo.AssertExpectations(s.T())
	s.mockTransactionSvc.AssertExpectations(s.T())
}

func (s *BillServiceTestSuite) TestConsolidateBills_NotHost() {
	roomId := "room1"
	userId := "2" // Not the host
	hostId := uint(1)

	room := &model.Room{ID: roomId, HostID: hostId}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockBillRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockBillRepo)
	s.mockTransactionRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockTransactionRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.billService.ConsolidateBills(roomId, userId)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrOnlyHostCanConsolidate, err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockBillRepo.AssertNotCalled(s.T(), "ConsolidateBills")
}

func (s *BillServiceTestSuite) TestConsolidateBills_AlreadyConsolidated() {
	roomId := "room1"
	userId := "1"
	hostId := uint(1)

	room := &model.Room{ID: roomId, HostID: hostId}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockBillRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockBillRepo)
	s.mockTransactionRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockTransactionRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.CONSOLIDATED, nil)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.billService.ConsolidateBills(roomId, userId)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrAlreadyConsolidated, err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockTransactionRepo.AssertNotCalled(s.T(), "Create")
}

func (s *BillServiceTestSuite) TestGetRoomBillConsolidationStatus_Unconsolidated_Success() {
	roomId := "room1"

	// Mock expectations
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.UNCONSOLIDATED, nil)

	// Execute
	status, err := s.billService.GetRoomBillConsolidationStatus(roomId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.True(s.T(), status == repository.UNCONSOLIDATED)
	s.mockBillRepo.AssertExpectations(s.T())
}
