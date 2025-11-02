package response

import "time"

type TransactionDto struct {
	ID              uint           `json:"id" binding:"required"`
	ConsolidationID uint           `json:"consolidationId" binding:"required"`
	Amount          float32        `json:"amount" binding:"required"`
	IsPaid          bool           `json:"isPaid" binding:"required"`
	PaidOn          time.Time      `json:"paidOn"`
	Payer           MinimalUserDto `json:"payer" binding:"required"`
	Payee           MinimalUserDto `json:"payee" binding:"required"`
}
