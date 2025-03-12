package test_services

import (
	"database/sql"
	"errors"
	"log"
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

type BillServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	billService *services.BillService
}

func (s *BillServiceTestSuite) SetupTest() {
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

	s.billService = &services.BillService{DB: s.DB}
}

func (s *BillServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *BillServiceTestSuite) TestCreateBill_Success() {
	// arrange
	now := time.Now()

	room := &model.Room{
		ID:        "room-123",
		Name:      "Test Room",
		CreatedAt: now,
	}

	owner := &model.User{
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

	payers := []model.User{payer1, payer2}
	name := "Dinner"
	amount := float32(100.0)
	includeOwner := true

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
		WillReturnRows(sqlmock.NewRows([]string{"id", "consolidation_id"}).AddRow(1, 1))

	s.mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(
			payer1.Username, payer1.Email, payer1.Password, payer1.PictureUrl,
			payer1.IsEmailValid, payer1.IsOnline, payer1.LastSeen,
			payer1.RegisteredAt, payer1.UpdatedAt, payer1.ID,
			payer2.Username, payer2.Email, payer2.Password, payer2.PictureUrl,
			payer2.IsEmailValid, payer2.IsOnline, payer2.LastSeen,
			payer2.RegisteredAt, payer2.UpdatedAt, payer2.ID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(payer1.ID).
			AddRow(payer2.ID))

	// Expect many-to-many relationship with payers
	s.mock.ExpectExec(`INSERT INTO "payers" \("bill_id","bill_room_id","bill_owner_id","user_id"\) VALUES \(\$1,\$2,\$3,\$4\),\(\$5,\$6,\$7,\$8\) ON CONFLICT DO NOTHING`).
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
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), bill)
	assert.Equal(s.T(), name, bill.Name)
	assert.Equal(s.T(), amount, bill.Amount)
	assert.Equal(s.T(), includeOwner, bill.IncludeOwner)
	assert.Equal(s.T(), room.ID, bill.RoomID)
	assert.Equal(s.T(), owner.ID, bill.OwnerID)
	assert.Equal(s.T(), 2, len(bill.Payers))
}

func (s *BillServiceTestSuite) TestCreateBill_EmptyPayers() {
	// arrange
	now := time.Now()

	room := &model.Room{
		ID:        "room-123",
		Name:      "Test Room",
		CreatedAt: now,
	}

	owner := &model.User{
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

	payers := []model.User{}
	name := "Dinner"
	amount := float32(100.0)
	includeOwner := true

	// act
	bill, err := s.billService.CreateBill(room, owner, name, amount, includeOwner, &payers)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), bill)
	assert.Contains(s.T(), err.Error(), "payers of a bill can't be empty")
}

func (s *BillServiceTestSuite) TestCreateBill_DatabaseError() {
	// arrange
	now := time.Now()

	room := &model.Room{
		ID:        "room-123",
		Name:      "Test Room",
		CreatedAt: now,
	}

	owner := &model.User{
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

	payers := []model.User{payer1}
	name := "Dinner"
	amount := float32(100.0)
	includeOwner := true

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
	assert.Error(s.T(), err)
	assert.Nil(s.T(), bill)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestGetBillById_Success() {
	// arrange
	now := time.Now()
	billId := uint(1)

	bill := model.Bill{
		ID:           billId,
		Name:         "Dinner",
		Amount:       100.0,
		Date:         now,
		IncludeOwner: true,
		RoomID:       "room-123",
		OwnerID:      1,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "amount", "date", "include_owner", "room_id", "owner_id", "consolidation_id",
	}).AddRow(
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
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
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
	now := time.Now()
	roomId := "room-123"

	bills := []model.Bill{
		{
			ID:           1,
			Name:         "Dinner",
			Amount:       100.0,
			Date:         now,
			IncludeOwner: true,
			RoomID:       roomId,
			OwnerID:      1,
		},
		{
			ID:           2,
			Name:         "Movie",
			Amount:       50.0,
			Date:         now,
			IncludeOwner: false,
			RoomID:       roomId,
			OwnerID:      2,
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "amount", "date", "include_owner", "room_id", "owner_id", "consolidation_id",
	})

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
		WithArgs(roomId).
		WillReturnRows(rows)

	// Expect preloading Owner
	ownerRows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(1, "owner1").AddRow(2, "owner2")
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnRows(ownerRows)

	// Expect preloading Payers for first bill
	payersJoinRows := sqlmock.NewRows([]string{"bill_id", "user_id", "bill_room_id", "bill_owner_id"}).
		AddRow(1, 3, roomId, 1).
		AddRow(1, 4, roomId, 1).
		AddRow(2, 5, roomId, 2).
		AddRow(2, 6, roomId, 2)
	s.mock.ExpectQuery(`SELECT \* FROM "payers" WHERE \("payers"."bill_id","payers"."bill_room_id","payers"."bill_owner_id"\) IN \(\(\$1,\$2,\$3\),\(\$4,\$5,\$6\)\)`).
		WithArgs(1, roomId, 1, 2, roomId, 2).
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
	result, err := s.billService.GetBillsForRoom(roomId)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), 2, len(*result))
	assert.Equal(s.T(), uint(1), (*result)[0].ID)
	assert.Equal(s.T(), uint(2), (*result)[1].ID)
}

func (s *BillServiceTestSuite) TestGetBillsForRoom_EmptyResult() {
	// arrange
	roomId := "room-123"

	rows := sqlmock.NewRows([]string{
		"id", "name", "amount", "date", "include_owner", "room_id", "owner_id", "consolidation_id",
	})

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1`).
		WithArgs(roomId).
		WillReturnRows(rows)

	// act
	result, err := s.billService.GetBillsForRoom(roomId)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), 0, len(*result))
}

func (s *BillServiceTestSuite) TestGetBillsForRoom_DatabaseError() {
	// arrange
	roomId := "room-123"

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1`).
		WithArgs(roomId).
		WillReturnError(errors.New("database error"))

	// act
	result, err := s.billService.GetBillsForRoom(roomId)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestDeleteRoomBills_Success() {
	// arrange
	roomId := "room-123"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "bills" WHERE room_id = \$1`).
		WithArgs(roomId).
		WillReturnResult(sqlmock.NewResult(0, 2)) // 2 rows affected
	s.mock.ExpectCommit()

	// act
	err := s.billService.DeleteRoomBills(roomId)

	// assert
	assert.NoError(s.T(), err)
}

func (s *BillServiceTestSuite) TestDeleteRoomBills_DatabaseError() {
	// arrange
	roomId := "room-123"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "bills" WHERE room_id = \$1`).
		WithArgs(roomId).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	err := s.billService.DeleteRoomBills(roomId)

	// assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestIsRoomBillConsolidated_True() {
	// arrange
	roomId := "room-123"
	consolidationId := uint(1)

	rows := sqlmock.NewRows([]string{
		"id", "consolidation_id",
	}).AddRow(
		1, consolidationId,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1 ORDER BY "bills"."id" LIMIT \$2`).
		WithArgs(roomId, 1).
		WillReturnRows(rows)

	// act
	result, err := s.billService.IsRoomBillConsolidated(roomId)

	// assert
	assert.NoError(s.T(), err)
	assert.True(s.T(), result)
}

func (s *BillServiceTestSuite) TestIsRoomBillConsolidated_False() {
	// arrange
	roomId := "room-123"
	var consolidationId uint = 0

	rows := sqlmock.NewRows([]string{
		"id", "consolidation_id",
	}).AddRow(
		1, consolidationId,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1 ORDER BY "bills"."id" LIMIT \$2`).
		WithArgs(roomId, 1).
		WillReturnRows(rows)

	// act
	result, err := s.billService.IsRoomBillConsolidated(roomId)

	// assert
	assert.NoError(s.T(), err)
	assert.False(s.T(), result)
}

func (s *BillServiceTestSuite) TestIsRoomBillConsolidated_DatabaseError() {
	// arrange
	roomId := "room-123"

	s.mock.ExpectQuery(`SELECT \* FROM "bills" WHERE room_id = \$1 ORDER BY "bills"."id" LIMIT \$2`).
		WithArgs(roomId, 1).
		WillReturnError(errors.New("database error"))

	// act
	result, err := s.billService.IsRoomBillConsolidated(roomId)

	// assert
	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestConsolidateBills_Success() {
	// arrange
	roomId := "room-123"
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
		WithArgs(consolidationId, roomId).
		WillReturnResult(sqlmock.NewResult(0, 2)) // 2 rows affected
	s.mock.ExpectCommit()

	// act
	result, err := s.billService.ConsolidateBills(roomId)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), consolidationId, result.ID)
}

func (s *BillServiceTestSuite) TestConsolidateBills_CreateError() {
	// arrange
	roomId := "room-123"

	// Expect creating the consolidation with error
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(`INSERT INTO "consolidations"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
		).
		WillReturnError(errors.New("database error"))

	// act
	result, err := s.billService.ConsolidateBills(roomId)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *BillServiceTestSuite) TestConsolidateBills_UpdateError() {
	// arrange
	roomId := "room-123"
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
		WithArgs(consolidationId, roomId).
		WillReturnError(errors.New("database error"))

	// act
	result, err := s.billService.ConsolidateBills(roomId)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func TestBillServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BillServiceTestSuite))
}
