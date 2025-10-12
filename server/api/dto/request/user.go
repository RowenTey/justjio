package request

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}
type ModifyFriendRequest struct {
	FriendID uint `json:"friendId"`
}

type RespondToFriendRequestRequest struct {
	Action    string `json:"action"`
	RequestID uint   `json:"requestId"`
}
