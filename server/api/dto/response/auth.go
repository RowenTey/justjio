package response

type AuthResponse struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	PictureUrl string `json:"pictureUrl"`
	UID        uint   `json:"id"`
}
