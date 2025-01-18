package services

import (
	"log"
	"time"

	"github.com/RowenTey/JustJio/model"

	"gorm.io/gorm"
)

type BillService struct {
	DB *gorm.DB
}

func (bs *BillService) CreateBill(
	room *model.Room,
	owner *model.User,
	name string,
	amount float32,
	payers *[]model.User,
) (*model.Bill, error) {
	// TODO: Modify to accept list of attendee IDs that should pay for this bill
	db := bs.DB.Table("bills")

	bill := model.Bill{
		Name:    name,
		Amount:  amount,
		Date:    time.Now(),
		Room:    *room,
		Owner:   *owner,
		OwnerID: owner.ID,
		Payers:  *payers,
	}

	// Omit to avoid creating new room
	if err := db.Omit("Room").Create(&bill).Error; err != nil {
		return nil, err
	}

	log.Println("[BILL] Bill created: ", bill.ID)
	return &bill, nil
}

func (bs *BillService) GetBillById(billId uint) (model.Bill, error) {
	db := bs.DB.Table("bills")
	var bill model.Bill

	if err := db.Where("id = ?", billId).First(&bill).Error; err != nil {
		return model.Bill{}, err
	}

	return bill, nil
}

func (bs *BillService) GetBillsForRoom(roomId string) (*[]model.Bill, error) {
	db := bs.DB.Table("bills")
	var bills []model.Bill

	if err := db.Where("room_id = ?", roomId).Find(&bills).Error; err != nil {
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

// Consolidate bills for a room
func (bs *BillService) ConsolidateBills(roomId string) (*model.Consolidation, error) {
	db := bs.DB.Table("bills")

	// Create empty struct as fields will be auto populated by DB
	consolidation := model.Consolidation{}
	if err := bs.DB.Table("consolidation").
		Create(&consolidation).Error; err != nil {
		return nil, err
	}
	log.Println("[BILL] Created consolidation of bills: ", consolidation.ID)

	if err := db.
		Where("room_id = ?", roomId).
		Update("consolidation_id", consolidation.ID).Error; err != nil {
		return nil, err
	}

	return &consolidation, nil
}
