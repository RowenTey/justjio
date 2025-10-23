package response

type AuthResponse struct {
	Username   string `json:"username" binding:"required"`
	Email      string `json:"email" binding:"required"`
	PictureUrl string `json:"pictureUrl" binding:"required"`
	UID        uint   `json:"id" binding:"required"`
}
