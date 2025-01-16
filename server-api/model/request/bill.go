package request

type CreateBillRequest struct {
	Name   string  `json:"name"`
	Amount float32 `json:"amount"`
	RoomID string  `json:"roomId"`
	Payers []uint  `json:"payers"`
}

type ConsolidateBillsRequest struct {
	RoomID string `json:"roomId"`
}
