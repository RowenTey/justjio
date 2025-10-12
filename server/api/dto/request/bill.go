package request

type CreateBillRequest struct {
	Name         string   `json:"name" validate:"required,min=1,max=100"`
	Amount       float32  `json:"amount" validate:"required,gt=0"`
	IncludeOwner bool     `json:"includeOwner"`
	RoomID       string   `json:"roomId" validate:"required"`
	Payers       []string `json:"payers" validate:"required,min=1,dive,required"`
}

type ConsolidateBillsRequest struct {
	RoomID string `json:"roomId" validate:"required"`
}
