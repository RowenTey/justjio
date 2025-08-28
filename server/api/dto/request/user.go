package request

type UpdateUserRequest struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type ModifyFriendRequest struct {
	FriendID uint `json:"friendId"`
}

type RespondToFriendRequestRequest struct {
	Action    string `json:"action"`
	RequestID uint   `json:"requestId"`
}
