package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/internal/repositories"
	"github.com/RowenTey/JustJio/server/api/pkg/database"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/response"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrAlreadyConsolidated    = errors.New("bills for this room have already been consolidated")
	ErrPayersNotFound         = errors.New("payer(s) not found")
	ErrOnlyHostCanConsolidate = errors.New("only the host can consolidate bills")
)

type BillService struct {
	db                 *gorm.DB
	billRepo           repositories.BillRepository
	userRepo           repositories.UserRepository
	roomRepo           repositories.RoomRepository
	transactionRepo    repositories.TransactionRepository
	transactionService TransactionService
	logger             *logrus.Entry
}

func NewBillService(
	db *gorm.DB,
	billRepo repositories.BillRepository,
	userRepo repositories.UserRepository,
	roomRepo repositories.RoomRepository,
	transactionRepo repositories.TransactionRepository,
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
		logger:             logger.WithFields(logrus.Fields{"service": "BillService"}),
	}
}

func (bs *BillService) CreateBill(
	ctx context.Context,
	roomId string,
	ownerid string,
	payersId []string,
	name string,
	amount float32,
	includeOwner bool,
) (uint, error) {
	var bill models.Bill

	if err := database.RunInTransaction(bs.db, sql.LevelDefault, func(tx *gorm.DB) error {
		roomRepoTx := bs.roomRepo.WithTx(tx)
		billRepoTx := bs.billRepo.WithTx(tx)
		userRepoTx := bs.userRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(ctx, roomId)
		if err != nil {
			return err
		}

		if room.Consolidated == "CONSOLIDATED" {
			return ErrAlreadyConsolidated
		}

		owner, err := userRepoTx.FindByID(ctx, ownerid)
		if err != nil {
			return err
		}

		payers, err := userRepoTx.FindByIDs(ctx, payersId)
		if err != nil {
			return err
		}

		bill = models.Bill{
			Name:         name,
			Amount:       amount,
			Date:         time.Now(),
			IncludeOwner: includeOwner,
			RoomID:       room.ID,
			OwnerID:      owner.ID,
			Payers:       payers,
		}
		if err := billRepoTx.Create(ctx, &bill); err != nil {
			return err
		}

		room.Consolidated = "UNCONSOLIDATED"
		if err := roomRepoTx.Update(ctx, room); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return 0, err
	}

	bs.logger.Info("Bill created in room: ", bill.RoomID)
	return bill.ID, nil
}

func (bs *BillService) GetBillById(ctx context.Context, billId uint) (*response.BillDto, error) {
	bill, err := bs.billRepo.FindByID(ctx, billId)
	if err != nil {
		return nil, err
	}

	payersDto := make([]response.MinimalUserDto, len(bill.Payers))
	for j, payer := range bill.Payers {
		payersDto[j] = response.MinimalUserDto{
			ID:         payer.ID,
			Username:   payer.Username,
			PictureUrl: payer.PictureUrl,
		}
	}

	return &response.BillDto{
		ID:              bill.ID,
		Name:            bill.Name,
		Amount:          bill.Amount,
		Date:            bill.Date,
		IncludeOwner:    bill.IncludeOwner,
		ConsolidationID: bill.ConsolidationID,
		Owner: response.MinimalUserDto{
			ID:         bill.Owner.ID,
			Username:   bill.Owner.Username,
			PictureUrl: bill.Owner.PictureUrl,
		},
		Payers: payersDto,
	}, nil
}

func (bs *BillService) GetBillsForRoom(ctx context.Context, roomId string) ([]response.BillDto, error) {
	bill, err := bs.billRepo.FindByRoom(ctx, roomId)
	if err != nil {
		return nil, err
	}

	billsDto := make([]response.BillDto, len(bill))
	for i, b := range bill {
		payersDto := make([]response.MinimalUserDto, len(b.Payers))
		for j, payer := range b.Payers {
			payersDto[j] = response.MinimalUserDto{
				ID:         payer.ID,
				Username:   payer.Username,
				PictureUrl: payer.PictureUrl,
			}
		}

		billsDto[i] = response.BillDto{
			ID:              b.ID,
			Name:            b.Name,
			Amount:          b.Amount,
			Date:            b.Date,
			IncludeOwner:    b.IncludeOwner,
			ConsolidationID: b.ConsolidationID,
			Owner: response.MinimalUserDto{
				ID:         b.Owner.ID,
				Username:   b.Owner.Username,
				PictureUrl: b.Owner.PictureUrl,
			},
			Payers: payersDto,
		}
	}

	return billsDto, nil
}

func (bs *BillService) DeleteRoomBills(ctx context.Context, roomId string) error {
	return bs.billRepo.DeleteByRoom(ctx, roomId)
}

func (bs *BillService) ConsolidateBills(ctx context.Context, roomId, userId string) error {
	return database.RunInTransaction(bs.db, sql.LevelDefault, func(tx *gorm.DB) error {
		roomRepoTx := bs.roomRepo.WithTx(tx)
		billRepoTx := bs.billRepo.WithTx(tx)
		transactionRepoTx := bs.transactionRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(ctx, roomId)
		if err != nil {
			return err
		}

		if utils.UIntToString(room.HostID) != userId {
			return ErrOnlyHostCanConsolidate
		}

		if room.Consolidated == "CONSOLIDATED" {
			return ErrAlreadyConsolidated
		}

		bs.logger.Info("Consolidating bills...")
		consolidation, err := billRepoTx.ConsolidateBills(ctx, roomId)
		if err != nil {
			return err
		}
		bs.logger.Info("Bills consolidated: ", consolidation.ID)

		bills, err := billRepoTx.FindByConsolidation(ctx, consolidation.ID)
		if err != nil {
			return err
		}

		transaction, err := bs.transactionService.GenerateTransactions(bills, consolidation)
		if err != nil {
			return err
		}

		if err := transactionRepoTx.Create(ctx, transaction); err != nil {
			return err
		}

		room.Consolidated = "CONSOLIDATED"
		if err := roomRepoTx.Update(ctx, room); err != nil {
			return err
		}

		bs.logger.Info("Created bills consolidation: ", consolidation.ID)
		return nil
	})
}
