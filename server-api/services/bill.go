package services

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/model"

	"gorm.io/gorm"
)

type BillService struct {
	DB     *gorm.DB
	logger *log.Entry
}

func NewBillService(db *gorm.DB) *BillService {
	return &BillService{
		DB:     db,
		logger: log.WithFields(log.Fields{"service": "BillService"}),
	}
}

func (bs *BillService) CreateBill(
	room *model.Room,
	owner *model.User,
	name string,
	amount float32,
	includeOwner bool,
	payers *[]model.User,
) (*model.Bill, error) {
	db := bs.DB.Table("bills")

	if len(*payers) == 0 {
		return nil, errors.New("payers of a bill can't be empty")
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

	// Omit to avoid creating new room and set consolidation to null
	if err := db.Omit("Room", "Owner").Create(&bill).Error; err != nil {
		return nil, err
	}

	bs.logger.Info("Bill created in room: ", bill.RoomID)
	return &bill, nil
}

func (bs *BillService) GetBillById(billId uint) (*model.Bill, error) {
	db := bs.DB.Table("bills")
	var bill model.Bill

	if err := db.Where("id = ?", billId).First(&bill).Error; err != nil {
		return &model.Bill{}, err
	}

	return &bill, nil
}

func (bs *BillService) GetBillsForRoom(roomId string) (*[]model.Bill, error) {
	db := bs.DB.Table("bills")
	var bills []model.Bill

	if err := db.Where("room_id = ?", roomId).Preload("Owner").Preload("Payers").Find(&bills).Error; err != nil {
		return nil, err
	}

	return &bills, nil
}

func (bs *BillService) DeleteRoomBills(roomId string) error {
	db := bs.DB.Table("bills")

	if err := db.Where("room_id = ?", roomId).Delete(&model.Bill{}).Error; err != nil {
		return err
	}

	return nil
}

func (bs *BillService) IsRoomBillConsolidated(roomId string) (bool, error) {
	db := bs.DB.Table("bills")
	var bill model.Bill

	if err := db.Where("room_id = ?", roomId).First(&bill).Error; err != nil {
		return false, err
	}

	return bill.ConsolidationID != 0, nil
}

// Consolidate bills for a room
func (bs *BillService) ConsolidateBills(roomId string) (*model.Consolidation, error) {
	db := bs.DB.Table("bills")

	// Create empty struct as fields will be auto populated by DB
	consolidation := model.Consolidation{}
	if err := bs.DB.Table("consolidations").
		Create(&consolidation).Error; err != nil {
		return nil, err
	}
	bs.logger.Info("Created bills consolidation: ", consolidation.ID)

	if err := db.
		Where("room_id = ?", roomId).
		Update("consolidation_id", consolidation.ID).Error; err != nil {
		return nil, err
	}

	return &consolidation, nil
}
