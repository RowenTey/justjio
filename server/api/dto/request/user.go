package request

type UpdateUsernameRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
}

type ModifyFriendRequest struct {
	FriendID uint `json:"friendId" validate:"required"`
}

type RespondToFriendRequestRequest struct {
	Action    string `json:"action" validate:"required,oneof=accept reject"`
	RequestID uint   `json:"requestId" validate:"required"`
}
