package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrAlreadyConsolidated    = errors.New("bills for this room have already been consolidated")
	ErrEmptyPayers            = errors.New("payers of a bill can't be empty")
	ErrPayersNotFound         = errors.New("payer(s) not found")
	ErrOnlyHostCanConsolidate = errors.New("only the host can consolidate bills")
)

type BillService struct {
	db                 *gorm.DB
	billRepo           repository.BillRepository
	userRepo           repository.UserRepository
	roomRepo           repository.RoomRepository
	transactionRepo    repository.TransactionRepository
	transactionService TransactionService
	logger             *logrus.Entry
}

func NewBillService(
	db *gorm.DB,
	billRepo repository.BillRepository,
	userRepo repository.UserRepository,
	roomRepo repository.RoomRepository,
	transactionRepo repository.TransactionRepository,
	transactionService TransactionService,
	logger *logrus.Logger,
) *BillService {
	return &BillService{
		db:                 db,
		billRepo:           billRepo,
		userRepo:           userRepo,
		roomRepo:           roomRepo,
		transactionRepo:    transactionRepo,
		transactionService: transactionService,
		logger:             utils.AddServiceField(logger, "BillService"),
	}
}

func (bs *BillService) CreateBill(
	roomId string,
	ownerid string,
	payersId *[]uint,
	name string,
	amount float32,
	includeOwner bool,
) (*model.Bill, error) {
	if status, err := bs.billRepo.GetRoomBillConsolidationStatus(roomId); err != nil {
		return nil, err
	} else if status == repository.CONSOLIDATED {
		return nil, ErrAlreadyConsolidated
	}

	if len(*payersId) == 0 {
		return nil, ErrEmptyPayers
	}

	room, err := bs.roomRepo.GetByID(roomId)
	if err != nil {
		return nil, err
	}

	owner, err := bs.userRepo.FindByID(ownerid)
	if err != nil {
		return nil, err
	}

	payers, err := bs.userRepo.FindByIDs(payersId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPayersNotFound
		}
		return nil, err
	}

	bill := model.Bill{
		Name:         name,
		Amount:       amount,
		Date:         time.Now(),
		IncludeOwner: includeOwner,
		RoomID:       room.ID,
		OwnerID:      owner.ID,
		Payers:       *payers,
	}
	if err := bs.billRepo.Create(&bill); err != nil {
		return nil, err
	}

	bs.logger.Info("Bill created in room: ", bill.RoomID)
	return &bill, nil
}

func (bs *BillService) GetBillById(billId uint) (*model.Bill, error) {
	return bs.billRepo.FindByID(billId)
}

func (bs *BillService) GetBillsForRoom(roomId string) (*[]model.Bill, error) {
	return bs.billRepo.FindByRoom(roomId)
}

func (bs *BillService) DeleteRoomBills(roomId string) error {
	return bs.billRepo.DeleteByRoom(roomId)
}

func (bs *BillService) GetRoomBillConsolidationStatus(roomId string) (repository.Status, error) {
	if _, err := bs.roomRepo.GetByID(roomId); err != nil {
		return repository.NO_BILLS, err
	}

	status, err := bs.billRepo.GetRoomBillConsolidationStatus(roomId)
	if err != nil {
		return repository.NO_BILLS, err
	}
	return status, nil
}

func (bs *BillService) ConsolidateBills(roomId, userId string) error {
	return database.RunInTransaction(bs.db, sql.LevelDefault, func(tx *gorm.DB) error {
		roomRepoTx := bs.roomRepo.WithTx(tx)
		billRepoTx := bs.billRepo.WithTx(tx)
		transactionRepoTx := bs.transactionRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(roomId)
		if err != nil {
			return err
		}

		if utils.UIntToString(room.HostID) != userId {
			return ErrOnlyHostCanConsolidate
		}

		if status, err := bs.billRepo.GetRoomBillConsolidationStatus(roomId); err != nil {
			return err
		} else if status == repository.CONSOLIDATED {
			return ErrAlreadyConsolidated
		}

		bs.logger.Info("Consolidating bills...")
		consolidation, err := billRepoTx.ConsolidateBills(roomId)
		if err != nil {
			return err
		}
		bs.logger.Info("Bills consolidated: ", consolidation.ID)

		bills, err := billRepoTx.FindByConsolidation(consolidation.ID)
		if err != nil {
			return err
		}

		transaction, err := bs.transactionService.GenerateTransactions(bills, consolidation)
		if err != nil {
			return err
		}

		if err := transactionRepoTx.Create(transaction); err != nil {
			return err
		}

		bs.logger.Info("Created bills consolidation: ", consolidation.ID)
		return nil
	})
}
