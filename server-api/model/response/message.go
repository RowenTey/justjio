package response

import "github.com/RowenTey/JustJio/model"

type GetMessagesResponse struct {
	Messages  []model.Message `json:"messages"`
	Page      int             `json:"page"`
	PageCount int             `json:"pageCount"`
}

type GetNumRoomInvitationsResponse struct {
	Count int `json:"count"`
}
