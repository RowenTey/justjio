package services

import (
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/model"
	"github.com/RowenTey/JustJio/util"
)

type BillServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	billService *BillService

	roomId            string
	billCols          []string
	consolidationCols []string
}

func TestBillServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BillServiceTestSuite))
}

func (s *BillServiceTestSuite) SetupTest() {
	var err error
	s.DB, s.mock, err = util.SetupTestDB()
	assert.NoError(s.T(), err)

	s.billService = NewBillService(s.DB)

	s.roomId = "room-123"
	s.billCols = []string{"id", "name", "amount", "date", "include_owner", "room_id", "owner_id", "consolidation_id"}
	s.consolidationCols = []string{"id", "consolidation_id"}
}

func (s *BillServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *BillServiceTestSuite) TestCreateBill_Success() {
	// arrange
	room := createTestRoom(s.roomId)
	owner := createTestUser(1, "owner1", "owner@example.com")
	payer1 := createTestUser(2, "payer1", "payer1@example.com")
	payer2 := createTestUser(3, "payer2", "payer2@example.com")
	payers := []model.User{*payer1, *payer2}
	name, amount, includeOwner := prepareBillDetails()

	// Expect the bill creation
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "bills"`).
		WithArgs(
			name,
			amount,
			sqlmock.AnyArg(), // Date
			includeOwner,
			room.ID,
			owner.ID,
		).
		WillReturnRows(sqlmock.NewRows(s.consolidationCols).AddRow(1, 1))

	args := append(userArgs(payer1), userArgs(payer2)...)
	s.mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(payer1.ID).
			AddRow(payer2.ID))

	// Expect many-to-many relationship with payers
	s.mock.ExpectExec(`INSERT INTO "payers"`).
		WithArgs(
			1, // Bill ID
			room.ID,
			owner.ID,
			payers[0].ID,
			1, // Bill ID
			room.ID,
			owner.ID,
			payers[1].ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 2)) // 2 rows affected
	s.mock.ExpectCommit()

	// act
	bill, err := s.billService.CreateBill(room, owner, name, amount, includeOwner, &payers)

	// assert
	util.AssertNoErrAndNotNil(s.T(), err, bill)
	assert.Equal(s.T(), name, bill.Name)
	assert.Equal(s.T(), amount, bill.Amount)
	assert.Equal(s.T(), includeOwner, bill.IncludeOwner)
	assert.Equal(s.T(), room.ID, bill.RoomID)
	assert.Equal(s.T(), owner.ID, bill.OwnerID)
	assert.Equal(s.T(), 2, len(bill.Payers))
}

func (s *BillServiceTestSuite) TestCreateBill_EmptyPayers() {
	// arrange
	room := createTestRoom(s.roomId)
	owner := createTestUser(1, "owner1", "owner@example.com")
	payers := []model.User{}
	name, amount, includeOwner := prepareBillDetails()

	// act
	bill, err := s.billService.CreateBill(room, owner, name, amount, includeOwner, &payers)

	// assert
	util.AssertErrAndNil(s.T(), err, bill)
	assert.Contains(s.T(), err.Error(), "payers of a bill can't be empty")
}

func (s *BillServiceTestSuite) TestCreateBill_DatabaseError() {
	// arrange
	room := createTestRoom(s.roomId)
	owner := createTestUser(1, "owner1", "owner@example.com")
	payer1 := createTestUser(2, "payer1", "payer1@example.com")
	payers := []model.User{*payer1}
	name, amount, includeOwner := prepareBillDetails()

	// Expect the bill creation with database error
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "bills"`).
		WithArgs(
			name,
			amount,
			sqlmock.AnyArg(), // Date
			includeOwner,
			room.ID,
			owner.ID,
		).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	bill, err := s.billService.CreateBill(room, owner, name, amount, includeOwner, &payers)

	// assert
	util.AssertErrAndNil(s.T(), err, bill)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestGetBillById_Success() {
	// arrange
	billId := uint(1)
	bill := createTestBill(billId, "Dinner", 100.0, s.roomId, 1)

	rows := sqlmock.NewRows(s.billCols).AddRow(
		bill.ID,
		bill.Name,
		bill.Amount,
		bill.Date,
		bill.IncludeOwner,
		bill.RoomID,
		bill.OwnerID,
		bill.ConsolidationID,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE id = \$1 ORDER BY "bills"."id" LIMIT \$2`).
		WithArgs(billId, 1).
		WillReturnRows(rows)

	// act
	result, err := s.billService.GetBillById(billId)

	// assert
	util.AssertNoErrAndNotNil(s.T(), err, result)
	assert.Equal(s.T(), billId, result.ID)
	assert.Equal(s.T(), "Dinner", result.Name)
	assert.Equal(s.T(), float32(100.0), result.Amount)
}

func (s *BillServiceTestSuite) TestGetBillById_NotFound() {
	// arrange
	billId := uint(999)

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE id = \$1 ORDER BY "bills"."id" LIMIT \$2`).
		WithArgs(billId, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// act
	result, err := s.billService.GetBillById(billId)

	// assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), uint(0), result.ID) // Empty bill returned
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}

func (s *BillServiceTestSuite) TestGetBillsForRoom_Success() {
	// arrange
	bills := []model.Bill{
		createTestBill(1, "Dinner", 100.0, s.roomId, 1),
		createTestBill(2, "Movie", 50.0, s.roomId, 2),
	}

	rows := sqlmock.NewRows(s.billCols)
	for _, bill := range bills {
		rows.AddRow(
			bill.ID,
			bill.Name,
			bill.Amount,
			bill.Date,
			bill.IncludeOwner,
			bill.RoomID,
			bill.OwnerID,
			bill.ConsolidationID,
		)
	}

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1`).
		WithArgs(s.roomId).
		WillReturnRows(rows)

	// Expect preloading Owner
	ownerRows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(1, "owner1").
		AddRow(2, "owner2")
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnRows(ownerRows)

	// Expect preloading Payers for first bill
	payersJoinRows := sqlmock.NewRows([]string{"bill_id", "user_id", "bill_room_id", "bill_owner_id"}).
		AddRow(1, 3, s.roomId, 1).
		AddRow(1, 4, s.roomId, 1).
		AddRow(2, 5, s.roomId, 2).
		AddRow(2, 6, s.roomId, 2)
	s.mock.ExpectQuery(`SELECT \* FROM "payers" WHERE \("payers"."bill_id","payers"."bill_room_id","payers"."bill_owner_id"\) IN \(\(\$1,\$2,\$3\),\(\$4,\$5,\$6\)\)`).
		WithArgs(1, s.roomId, 1, 2, s.roomId, 2).
		WillReturnRows(payersJoinRows)

	payerRows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(3, "payer1").
		AddRow(4, "payer2").
		AddRow(5, "payer3").
		AddRow(6, "payer4")
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2\,\$3,\$4\)`).
		WithArgs(3, 4, 5, 6).
		WillReturnRows(payerRows)

	// act
	result, err := s.billService.GetBillsForRoom(s.roomId)

	// assert
	util.AssertNoErrAndNotNil(s.T(), err, result)
	assert.Equal(s.T(), 2, len(*result))
	assert.Equal(s.T(), uint(1), (*result)[0].ID)
	assert.Equal(s.T(), uint(2), (*result)[1].ID)
}

func (s *BillServiceTestSuite) TestGetBillsForRoom_EmptyResult() {
	// arrange
	rows := sqlmock.NewRows(s.billCols)

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1`).
		WithArgs(s.roomId).
		WillReturnRows(rows)

	// act
	result, err := s.billService.GetBillsForRoom(s.roomId)

	// assert
	util.AssertNoErrAndNotNil(s.T(), err, result)
	assert.Equal(s.T(), 0, len(*result))
}

func (s *BillServiceTestSuite) TestGetBillsForRoom_DatabaseError() {
	// arrange
	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1`).
		WithArgs(s.roomId).
		WillReturnError(errors.New("database error"))

	// act
	result, err := s.billService.GetBillsForRoom(s.roomId)

	// assert
	util.AssertErrAndNil(s.T(), err, result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestDeleteRoomBills_Success() {
	// arrange
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "bills" WHERE room_id = \$1`).
		WithArgs(s.roomId).
		WillReturnResult(sqlmock.NewResult(0, 2)) // 2 rows affected
	s.mock.ExpectCommit()

	// act
	err := s.billService.DeleteRoomBills(s.roomId)

	// assert
	assert.NoError(s.T(), err)
}

func (s *BillServiceTestSuite) TestDeleteRoomBills_DatabaseError() {
	// arrange
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "bills" WHERE room_id = \$1`).
		WithArgs(s.roomId).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	err := s.billService.DeleteRoomBills(s.roomId)

	// assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestIsRoomBillConsolidated_True() {
	// arrange
	consolidationId := uint(1)
	rows := sqlmock.NewRows(s.consolidationCols).AddRow(
		1, consolidationId,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1 ORDER BY "bills"."id" LIMIT \$2`).
		WithArgs(s.roomId, 1).
		WillReturnRows(rows)

	// act
	result, err := s.billService.IsRoomBillConsolidated(s.roomId)

	// assert
	assert.NoError(s.T(), err)
	assert.True(s.T(), result)
}

func (s *BillServiceTestSuite) TestIsRoomBillConsolidated_False() {
	// arrange
	consolidationId := uint(0)
	rows := sqlmock.NewRows(s.consolidationCols).AddRow(
		1, consolidationId,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1 ORDER BY "bills"."id" LIMIT \$2`).
		WithArgs(s.roomId, 1).
		WillReturnRows(rows)

	// act
	result, err := s.billService.IsRoomBillConsolidated(s.roomId)

	// assert
	assert.NoError(s.T(), err)
	assert.False(s.T(), result)
}

func (s *BillServiceTestSuite) TestIsRoomBillConsolidated_DatabaseError() {
	// arrange
	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1 ORDER BY "bills"."id" LIMIT \$2`).
		WithArgs(s.roomId, 1).
		WillReturnError(errors.New("database error"))

	// act
	result, err := s.billService.IsRoomBillConsolidated(s.roomId)

	// assert
	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestConsolidateBills_Success() {
	// arrange
	consolidationId := uint(1)

	// Expect creating the consolidation
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "consolidations"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(consolidationId))
	s.mock.ExpectCommit()

	// Expect updating the bills
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "bills" SET "consolidation_id"=\$1 WHERE room_id = \$2`).
		WithArgs(consolidationId, s.roomId).
		WillReturnResult(sqlmock.NewResult(0, 2)) // 2 rows affected
	s.mock.ExpectCommit()

	// act
	result, err := s.billService.ConsolidateBills(s.roomId)

	// assert
	util.AssertNoErrAndNotNil(s.T(), err, result)
	assert.Equal(s.T(), consolidationId, result.ID)
}

func (s *BillServiceTestSuite) TestConsolidateBills_CreateError() {
	// arrange
	// Expect creating the consolidation with error
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "consolidations"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
		).
		WillReturnError(errors.New("database error"))

	// act
	result, err := s.billService.ConsolidateBills(s.roomId)

	// assert
	util.AssertErrAndNil(s.T(), err, result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestConsolidateBills_UpdateError() {
	// arrange
	consolidationId := uint(1)

	// Expect creating the consolidation
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "consolidations"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(consolidationId))
	s.mock.ExpectCommit()

	// Expect updating the bills with error
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "bills" SET "consolidation_id"=\$1 WHERE room_id = \$2`).
		WithArgs(consolidationId, s.roomId).
		WillReturnError(errors.New("database error"))

	// act
	result, err := s.billService.ConsolidateBills(s.roomId)

	// assert
	util.AssertErrAndNil(s.T(), err, result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func createTestRoom(id string) *model.Room {
	now := time.Now()
	return &model.Room{
		ID:        id,
		Name:      "Test Room",
		CreatedAt: now,
	}
}

func createTestUser(id uint, username, email string) *model.User {
	now := time.Now()
	return &model.User{
		ID:           id,
		Username:     username,
		Email:        email,
		Password:     "password",
		PictureUrl:   "https://example.com/pic.jpg",
		IsEmailValid: true,
		IsOnline:     true,
		LastSeen:     now,
		RegisteredAt: now,
		UpdatedAt:    now,
	}
}

func createTestBill(id uint, name string, amount float32, roomId string, ownerId uint) model.Bill {
	now := time.Now()
	return model.Bill{
		ID:           id,
		Name:         name,
		Amount:       amount,
		Date:         now,
		IncludeOwner: true,
		RoomID:       roomId,
		OwnerID:      ownerId,
	}
}

func userArgs(u *model.User) []driver.Value {
	return []driver.Value{
		u.Username, u.Email, u.Password, u.PictureUrl,
		u.IsEmailValid, u.IsOnline, u.LastSeen,
		u.RegisteredAt, u.UpdatedAt, u.ID,
	}
}

func prepareBillDetails() (string, float32, bool) {
	name := "Dinner"
	amount := float32(100.0)
	includeOwner := true
	return name, amount, includeOwner
}
