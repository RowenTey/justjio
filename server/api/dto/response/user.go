package response

type GetNumFriendsResponse struct {
	NumFriends int64 `json:"numFriends"`
}

type CountPendingRequestsResponse struct {
	Count int64 `json:"count"`
}
