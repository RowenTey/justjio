package request

type UpdateUserRequest struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type ModifyFriendRequest struct {
	FriendID string `json:"friendId"`
}
