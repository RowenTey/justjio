package test_services

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"
	"math"
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

type floatMatcher struct {
	expected float64
	epsilon  float64
}

func (m floatMatcher) Match(v driver.Value) bool {
	actual, ok := v.(float64)
	if !ok {
		return false
	}
	return math.Abs(actual-m.expected) < m.epsilon
}

type TransactionServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	transactionService *services.TransactionService
}

func (s *TransactionServiceTestSuite) SetupTest() {
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

	s.transactionService = &services.TransactionService{DB: s.DB}
}

func (s *TransactionServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *TransactionServiceTestSuite) TestGenerateTransactions_Success() {
	// arrange
	now := time.Now()
	consolidatedBill := &model.Consolidation{
		ID:        1,
		CreatedAt: now,
	}

	// Setup mock data for bills with more complete User objects
	owner := model.User{
		ID:           1,
		Username:     "owner1",
		Email:        "owner@example.com",
		Password:     "password",
		PictureUrl:   "https://example.com/pic1.jpg",
		IsEmailValid: true,
		IsOnline:     true,
		LastSeen:     now,
		RegisteredAt: now,
		UpdatedAt:    now,
	}

	payer1 := model.User{
		ID:           2,
		Username:     "payer1",
		Email:        "payer1@example.com",
		Password:     "password",
		PictureUrl:   "https://example.com/pic2.jpg",
		IsEmailValid: true,
		IsOnline:     true,
		LastSeen:     now,
		RegisteredAt: now,
		UpdatedAt:    now,
	}

	payer2 := model.User{
		ID:           3,
		Username:     "payer2",
		Email:        "payer2@example.com",
		Password:     "password",
		PictureUrl:   "https://example.com/pic3.jpg",
		IsEmailValid: true,
		IsOnline:     true,
		LastSeen:     now,
		RegisteredAt: now,
		UpdatedAt:    now,
	}

	bills := []model.Bill{
		{
			ID:              1,
			Name:            "Dinner",
			Amount:          100.0,
			Date:            now,
			IncludeOwner:    true,
			RoomID:          "room-123",
			OwnerID:         owner.ID,
			ConsolidationID: consolidatedBill.ID,
			Owner:           owner,
			Payers:          []model.User{payer1, payer2},
		},
	}

	// Setup the query expectation for retrieving bills
	billsRows := sqlmock.NewRows([]string{"id", "name", "amount", "date", "include_owner", "room_id", "owner_id", "consolidation_id"}).
		AddRow(bills[0].ID, bills[0].Name, bills[0].Amount, bills[0].Date, bills[0].IncludeOwner, bills[0].RoomID, bills[0].OwnerID, bills[0].ConsolidationID)

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE consolidation_id = \$1`).
		WithArgs(consolidatedBill.ID).
		WillReturnRows(billsRows)

	// Setup expectations for preloading payers
	payersJoinRows := sqlmock.NewRows([]string{"bill_id", "user_id", "bill_room_id", "bill_owner_id"}).
		AddRow(bills[0].ID, bills[0].Payers[0].ID, bills[0].RoomID, bills[0].OwnerID).
		AddRow(bills[0].ID, bills[0].Payers[1].ID, bills[0].RoomID, bills[0].OwnerID)
	s.mock.ExpectQuery(`SELECT \* FROM "payers" WHERE \("payers"."bill_id","payers"."bill_room_id","payers"."bill_owner_id"\) IN \(\(\$1,\$2,\$3\)\)`).
		WithArgs(bills[0].ID, bills[0].RoomID, bills[0].OwnerID).
		WillReturnRows(payersJoinRows)

	// Return all fields for the payer users
	payerRows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "picture_url", "is_email_valid",
		"is_online", "last_seen", "registered_at", "updated_at",
	}).
		AddRow(
			payer1.ID, payer1.Username, payer1.Email, payer1.Password,
			payer1.PictureUrl, payer1.IsEmailValid, payer1.IsOnline,
			payer1.LastSeen, payer1.RegisteredAt, payer1.UpdatedAt,
		).
		AddRow(
			payer2.ID, payer2.Username, payer2.Email, payer2.Password,
			payer2.PictureUrl, payer2.IsEmailValid, payer2.IsOnline,
			payer2.LastSeen, payer2.RegisteredAt, payer2.UpdatedAt,
		)
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2\)`).
		WithArgs(payer1.ID, payer2.ID).
		WillReturnRows(payerRows)

	// Expect the transaction creation
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "transactions"`).
		WithArgs(
			consolidatedBill.ID, // 1st row consolidation_id
			payer1.ID,           // 1st row payer_id
			owner.ID,            // 1st row payee_id
			floatMatcher{expected: 33.33, epsilon: 0.01}, // 1st row amount
			false,               // 1st row is_paid
			consolidatedBill.ID, // 2nd row consolidation_id
			payer2.ID,           // 2nd row payer_id
			owner.ID,            // 2nd row payee_id
			floatMatcher{expected: 33.33, epsilon: 0.01}, // 2nd row amount
			false, // 2nd row is_paid
		).
		WillReturnRows(sqlmock.NewRows([]string{"paid_on", "id"}).
			AddRow(time.Time{}, 1).
			AddRow(time.Time{}, 2))
	s.mock.ExpectCommit()

	// act
	err := s.transactionService.GenerateTransactions(consolidatedBill)

	// assert
	assert.NoError(s.T(), err)
}

func (s *TransactionServiceTestSuite) TestGenerateTransactions_DatabaseError() {
	// arrange
	consolidatedBill := &model.Consolidation{
		ID:        1,
		CreatedAt: time.Now(),
	}

	// Setup the query expectation for retrieving bills - return an error
	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE consolidation_id = \$1`).
		WithArgs(consolidatedBill.ID).
		WillReturnError(errors.New("database error"))

	// act
	err := s.transactionService.GenerateTransactions(consolidatedBill)

	// assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *TransactionServiceTestSuite) TestGetTransactionsByUser_Success() {
	// arrange
	isPaid := false
	userId := "1"

	expectedTransactions := []model.Transaction{
		{
			ID:              1,
			ConsolidationID: 1,
			PayerID:         2,
			PayeeID:         1,
			Amount:          50.0,
			IsPaid:          isPaid,
			PaidOn:          time.Time{},
		},
		{
			ID:              2,
			ConsolidationID: 1,
			PayerID:         1,
			PayeeID:         3,
			Amount:          30.0,
			IsPaid:          isPaid,
			PaidOn:          time.Time{},
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "consolidation_id", "payer_id", "payee_id", "amount", "is_paid", "paid_on",
	})

	for _, tx := range expectedTransactions {
		rows.AddRow(
			tx.ID,
			tx.ConsolidationID,
			tx.PayerID,
			tx.PayeeID,
			tx.Amount,
			tx.IsPaid,
			tx.PaidOn,
		)
	}

	s.mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE is_paid = \$1 AND \(payee_id = \$2 OR payer_id = \$3\)`).
		WithArgs(isPaid, userId, userId).
		WillReturnRows(rows)

	// Setup expectations for preloading payee
	payeeRows1 := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(1, "user1")
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2\)`).
		WithArgs(1, 3).
		WillReturnRows(payeeRows1)

	// Setup expectations for preloading payer for first transaction
	payerRows1 := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(2, "user2")
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2\)`).
		WithArgs(2, 1).
		WillReturnRows(payerRows1)

	// act
	transactions, err := s.transactionService.GetTransactionsByUser(isPaid, userId)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), transactions)
	assert.Equal(s.T(), 2, len(*transactions))
	assert.Equal(s.T(), uint(1), (*transactions)[0].ID)
	assert.Equal(s.T(), uint(2), (*transactions)[1].ID)
}

func (s *TransactionServiceTestSuite) TestGetTransactionsByUser_EmptyResult() {
	// arrange
	isPaid := true
	userId := "1"

	rows := sqlmock.NewRows([]string{
		"id", "consolidation_id", "payer_id", "payee_id", "amount", "is_paid", "paid_on",
	})

	s.mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE is_paid = \$1 AND \(payee_id = \$2 OR payer_id = \$3\)`).
		WithArgs(isPaid, userId, userId).
		WillReturnRows(rows)

	// act
	transactions, err := s.transactionService.GetTransactionsByUser(isPaid, userId)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), transactions)
	assert.Equal(s.T(), 0, len(*transactions))
}

func (s *TransactionServiceTestSuite) TestGetTransactionsByUser_DatabaseError() {
	// arrange
	isPaid := false
	userId := "1"

	s.mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE is_paid = \$1 AND \(payee_id = \$2 OR payer_id = \$3\)`).
		WithArgs(isPaid, userId, userId).
		WillReturnError(errors.New("database error"))

	// act
	transactions, err := s.transactionService.GetTransactionsByUser(isPaid, userId)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), transactions)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_Success() {
	// arrange
	transactionId := "1"
	userId := "2"
	transaction := model.Transaction{
		ID:              1,
		ConsolidationID: 1,
		PayerID:         2,
		PayeeID:         1,
		Amount:          50.0,
		IsPaid:          false,
	}

	rows := sqlmock.NewRows([]string{
		"id", "consolidation_id", "payer_id", "payee_id", "amount", "is_paid",
	}).AddRow(
		transaction.ID,
		transaction.ConsolidationID,
		transaction.PayerID,
		transaction.PayeeID,
		transaction.Amount,
		transaction.IsPaid,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "transactions"."id" = \$1 ORDER BY "transactions"."id" LIMIT \$2`).
		WithArgs(transactionId, 1).
		WillReturnRows(rows)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "transactions" SET "payer_id"=\$1,"payee_id"=\$2,"amount"=\$3,"is_paid"=\$4,"paid_on"=\$5 WHERE "id" = \$6 AND "consolidation_id" = \$7`).
		WithArgs(transaction.PayerID, transaction.PayeeID, transaction.Amount, true, sqlmock.AnyArg(), transaction.ID, transaction.ConsolidationID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	result, err := s.transactionService.SettleTransaction(transactionId, userId)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), uint(1), result.ID)
	assert.True(s.T(), result.IsPaid)
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_NotFound() {
	// arrange
	transactionId := "999"
	userId := "2"

	s.mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "transactions"."id" = \$1 ORDER BY "transactions"."id" LIMIT \$2`).
		WithArgs(transactionId, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	result, err := s.transactionService.SettleTransaction(transactionId, userId)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_AlreadySettled() {
	// arrange
	transactionId := "1"
	userId := "2"
	transaction := model.Transaction{
		ID:              1,
		ConsolidationID: 1,
		PayerID:         2,
		PayeeID:         1,
		Amount:          50.0,
		IsPaid:          true, // Already settled
	}

	rows := sqlmock.NewRows([]string{
		"id", "consolidation_id", "payer_id", "payee_id", "amount", "is_paid",
	}).AddRow(
		transaction.ID,
		transaction.ConsolidationID,
		transaction.PayerID,
		transaction.PayeeID,
		transaction.Amount,
		transaction.IsPaid,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "transactions"."id" = \$1 ORDER BY "transactions"."id" LIMIT \$2`).
		WithArgs(transactionId, 1).
		WillReturnRows(rows)

	// act
	result, err := s.transactionService.SettleTransaction(transactionId, userId)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "transaction already settled")
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_InvalidPayer() {
	// arrange
	transactionId := "1"
	userId := "3" // Different from PayerID
	transaction := model.Transaction{
		ID:              1,
		ConsolidationID: 1,
		PayerID:         2, // Different from userId
		PayeeID:         1,
		Amount:          50.0,
		IsPaid:          false,
	}

	rows := sqlmock.NewRows([]string{
		"id", "consolidation_id", "payer_id", "payee_id", "amount", "is_paid",
	}).AddRow(
		transaction.ID,
		transaction.ConsolidationID,
		transaction.PayerID,
		transaction.PayeeID,
		transaction.Amount,
		transaction.IsPaid,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "transactions"."id" = \$1 ORDER BY "transactions"."id" LIMIT \$2`).
		WithArgs(transactionId, 1).
		WillReturnRows(rows)

	// act
	result, err := s.transactionService.SettleTransaction(transactionId, userId)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "invalid payer")
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_DatabaseError() {
	// arrange
	transactionId := "1"
	userId := "2"
	transaction := model.Transaction{
		ID:              1,
		ConsolidationID: 1,
		PayerID:         2,
		PayeeID:         1,
		Amount:          50.0,
		IsPaid:          false,
	}

	rows := sqlmock.NewRows([]string{
		"id", "consolidation_id", "payer_id", "payee_id", "amount", "is_paid",
	}).AddRow(
		transaction.ID,
		transaction.ConsolidationID,
		transaction.PayerID,
		transaction.PayeeID,
		transaction.Amount,
		transaction.IsPaid,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "transactions"."id" = \$1 ORDER BY "transactions"."id" LIMIT \$2`).
		WithArgs(transactionId, 1).
		WillReturnRows(rows)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "transactions" SET "payer_id"=\$1,"payee_id"=\$2,"amount"=\$3,"is_paid"=\$4,"paid_on"=\$5 WHERE "id" = \$6 AND "consolidation_id" = \$7`).
		WithArgs(transaction.PayerID, transaction.PayeeID, transaction.Amount, true, sqlmock.AnyArg(), transaction.ID, transaction.ConsolidationID).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	result, err := s.transactionService.SettleTransaction(transactionId, userId)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func TestTransactionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionServiceTestSuite))
}
