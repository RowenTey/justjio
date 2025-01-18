package response

import "github.com/RowenTey/JustJio/model"

type GetMessagesResponse struct {
	Messages  []model.Message `json:"messages"`
	Page      int             `json:"page"`
	PageCount int             `json:"page_count"`
}
