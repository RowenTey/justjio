package response

import "time"

type BillDto struct {
	ID              uint             `json:"id" binding:"required"`
	Name            string           `json:"name" binding:"required"`
	Amount          float32          `json:"amount" binding:"required"`
	Date            time.Time        `json:"date" binding:"required"`
	IncludeOwner    bool             `json:"includeOwner" binding:"required"`
	ConsolidationID uint             `json:"consolidationId"`
	Owner           MinimalUserDto   `json:"owner" binding:"required"`
	Payers          []MinimalUserDto `json:"payers" binding:"required"`
}
