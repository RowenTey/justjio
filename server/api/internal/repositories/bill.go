package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"gorm.io/gorm"
)

type billRepoDto struct {
	ID              uint
	Name            string
	Amount          float32
	Date            time.Time
	IncludeOwner    bool
	RoomID          string
	OwnerID         uint
	ConsolidationID sql.NullInt64
	OwnerUsername   sql.NullString
	OwnerPictureURL sql.NullString
	PayerID         sql.NullInt64
	PayerUsername   sql.NullString
	PayerPictureURL sql.NullString
}

type BillRepository interface {
	WithTx(tx *gorm.DB) BillRepository

	Create(ctx context.Context, bill *models.Bill) error
	FindByID(ctx context.Context, billID uint) (*models.Bill, error)
	FindByRoom(ctx context.Context, roomID string) ([]models.Bill, error)
	FindByConsolidation(ctx context.Context, consolidationID uint) ([]models.Bill, error)
	DeleteByRoom(ctx context.Context, roomID string) error
	ConsolidateBills(ctx context.Context, roomID string) (*models.Consolidation, error)
}

type billRepository struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) BillRepository {
	return &billRepository{db: db}
}

// WithTx returns a new BillRepository with the provided transaction
func (r *billRepository) WithTx(tx *gorm.DB) BillRepository {
	if tx == nil {
		return r
	}
	return &billRepository{db: tx}
}

func (r *billRepository) Create(ctx context.Context, bill *models.Bill) error {
	return r.db.WithContext(ctx).Create(bill).Error
}

func (r *billRepository) FindByID(ctx context.Context, billID uint) (*models.Bill, error) {
	rows, err := r.db.WithContext(ctx).
		Table("bills b").
		Select(`
            b.*,
            owner.id              AS owner_id_,
            owner.username        AS owner_username,
            owner.picture_url     AS owner_picture_url,
            payer.id              AS payer_id,
            payer.username        AS payer_username,
            payer.picture_url     AS payer_picture_url
        `).
		Where("b.id = ?", billID).
		Joins("LEFT JOIN users owner ON b.owner_id = owner.id").
		Joins("LEFT JOIN payers p ON b.id = p.bill_id").
		Joins("LEFT JOIN users payer ON p.user_id = payer.id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println("Error closing rows: ", err)
		}
	}()

	var bill *models.Bill
	seenPayers := make(map[uint]struct{})

	for rows.Next() {
		var sc billRepoDto
		if err := r.db.ScanRows(rows, &sc); err != nil {
			return nil, err
		}

		if bill == nil {
			bill = &models.Bill{
				ID:           sc.ID,
				Name:         sc.Name,
				Amount:       sc.Amount,
				Date:         sc.Date,
				IncludeOwner: sc.IncludeOwner,
				RoomID:       sc.RoomID,
				OwnerID:      sc.OwnerID,
			}

			if sc.ConsolidationID.Valid {
				bill.ConsolidationID = uint(sc.ConsolidationID.Int64)
			}

			if sc.OwnerUsername.Valid {
				bill.Owner = models.User{
					ID:         sc.OwnerID,
					Username:   sc.OwnerUsername.String,
					PictureUrl: sc.OwnerPictureURL.String,
				}
			}
		}

		if sc.PayerID.Valid {
			uid := uint(sc.PayerID.Int64)
			if _, seen := seenPayers[uid]; !seen {
				seenPayers[uid] = struct{}{}
				bill.Payers = append(bill.Payers, models.User{
					ID:         uid,
					Username:   sc.PayerUsername.String,
					PictureUrl: sc.PayerPictureURL.String,
				})
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	} else if bill == nil {
		return nil, gorm.ErrRecordNotFound
	}

	return bill, nil
}

func (r *billRepository) FindByRoom(ctx context.Context, roomID string) ([]models.Bill, error) {
	rows, err := r.db.WithContext(ctx).
		Table("bills b").
		Select(`
            b.*,
            owner.id              AS owner_id_,
            owner.username        AS owner_username,
            owner.picture_url     AS owner_picture_url,
            payer.id              AS payer_id,
            payer.username        AS payer_username,
            payer.picture_url     AS payer_picture_url
        `).
		Where("b.room_id = ?", roomID).
		Joins("LEFT JOIN users owner ON b.owner_id = owner.id").
		Joins("LEFT JOIN payers p ON b.id = p.bill_id").
		Joins("LEFT JOIN users payer ON p.user_id = payer.id").
		Order("b.date DESC, b.id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println("Error closing rows: ", err)
		}
	}()

	billMap := make(map[uint]*models.Bill)
	payerSeen := make(map[uint]map[uint]struct{}) // billID → userID

	for rows.Next() {
		var sc billRepoDto
		if err := r.db.ScanRows(rows, &sc); err != nil {
			return nil, err
		}

		bill, exists := billMap[sc.ID]
		if !exists {
			bill = &models.Bill{
				ID:           sc.ID,
				Name:         sc.Name,
				Amount:       sc.Amount,
				Date:         sc.Date,
				IncludeOwner: sc.IncludeOwner,
				RoomID:       sc.RoomID,
				OwnerID:      sc.OwnerID,
			}

			if sc.ConsolidationID.Valid {
				bill.ConsolidationID = uint(sc.ConsolidationID.Int64)
			}

			if sc.OwnerUsername.Valid {
				bill.Owner = models.User{
					ID:         sc.OwnerID,
					Username:   sc.OwnerUsername.String,
					PictureUrl: sc.OwnerPictureURL.String,
				}
			}

			billMap[sc.ID] = bill
		}

		if sc.PayerID.Valid {
			uid := uint(sc.PayerID.Int64)
			if payerSeen[sc.ID] == nil {
				payerSeen[sc.ID] = make(map[uint]struct{})
			}

			if _, seen := payerSeen[sc.ID][uid]; !seen {
				payerSeen[sc.ID][uid] = struct{}{}
				bill.Payers = append(bill.Payers, models.User{
					ID:         uid,
					Username:   sc.PayerUsername.String,
					PictureUrl: sc.PayerPictureURL.String,
				})
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]models.Bill, 0, len(billMap))
	for _, b := range billMap {
		result = append(result, *b)
	}

	return result, nil
}

func (r *billRepository) DeleteByRoom(ctx context.Context, roomID string) error {
	return r.db.WithContext(ctx).Where("room_id = ?", roomID).Delete(&models.Bill{}).Error
}

func (r *billRepository) ConsolidateBills(ctx context.Context, roomID string) (*models.Consolidation, error) {
	// Create empty struct as fields will be auto populated by DB
	consolidation := models.Consolidation{}

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.db.
			WithContext(ctx).
			Model(&models.Consolidation{}).
			Create(&consolidation).Error; err != nil {
			return err
		}

		if err := r.db.WithContext(ctx).Table("bills").
			Where("room_id = ?", roomID).
			Update("consolidation_id", consolidation.ID).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &consolidation, nil
}

func (r *billRepository) FindByConsolidation(ctx context.Context, consolidationID uint) ([]models.Bill, error) {
	rows, err := r.db.WithContext(ctx).
		Table("bills b").
		Select(`
            b.*,
            owner.username        AS owner_username,
            owner.picture_url     AS owner_picture_url,
            payer.id              AS payer_id,
            payer.username        AS payer_username,
            payer.picture_url     AS payer_picture_url
        `).
		Where("b.consolidation_id = ?", consolidationID).
		Joins("LEFT JOIN users owner ON b.owner_id = owner.id").
		Joins("LEFT JOIN payers p ON b.id = p.bill_id").
		Joins("LEFT JOIN users payer ON p.user_id = payer.id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println("Error closing rows: ", err)
		}
	}()

	// deduplicate bills
	billMap := make(map[uint]*models.Bill)
	// bill_id → user_id
	payerSeen := make(map[uint]map[uint]struct{})

	for rows.Next() {
		var sc billRepoDto
		if err := r.db.ScanRows(rows, &sc); err != nil {
			return nil, err
		}

		bill, exists := billMap[sc.ID]
		if !exists {
			bill = &models.Bill{
				ID:           sc.ID,
				Name:         sc.Name,
				Amount:       sc.Amount,
				Date:         sc.Date,
				IncludeOwner: sc.IncludeOwner,
				RoomID:       sc.RoomID,
				OwnerID:      sc.OwnerID,
			}

			if sc.ConsolidationID.Valid {
				cid := uint(sc.ConsolidationID.Int64)
				bill.ConsolidationID = cid
			}

			if sc.OwnerUsername.Valid {
				bill.Owner = models.User{
					ID:         sc.OwnerID,
					Username:   sc.OwnerUsername.String,
					PictureUrl: sc.OwnerPictureURL.String,
				}
			}

			billMap[sc.ID] = bill
		}

		if sc.PayerID.Valid {
			billID := sc.ID
			userID := uint(sc.PayerID.Int64)

			if payerSeen[billID] == nil {
				payerSeen[billID] = make(map[uint]struct{})
			}

			if _, seen := payerSeen[billID][userID]; !seen {
				payerSeen[billID][userID] = struct{}{}
				bill.Payers = append(bill.Payers, models.User{
					ID:         userID,
					Username:   sc.PayerUsername.String,
					PictureUrl: sc.PayerPictureURL.String,
				})
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice
	result := make([]models.Bill, 0, len(billMap))
	for _, bill := range billMap {
		result = append(result, *bill)
	}

	return result, nil
}
