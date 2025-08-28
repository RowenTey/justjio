package response

type IsFriendResponse struct {
	IsFriend bool `json:"isFriend"`
}

type GetNumFriendsResponse struct {
	NumFriends int64 `json:"numFriends"`
}

type CountPendingRequestsResponse struct {
	Count int64 `json:"count"`
}
