package model

// Friend request status constants
const (
	FriendRequestStatusPending  = "pending"
	FriendRequestStatusAccepted = "accepted"
	FriendRequestStatusRejected = "rejected"
)

// Room invite status constants
const (
	RoomInviteStatusPending  = "pending"
	RoomInviteStatusAccepted = "accepted"
	RoomInviteStatusRejected = "rejected"
)

// OTP purpose constants
const (
	OTPPurposeVerifyEmail   = "verify-email"
	OTPPurposeResetPassword = "reset-password"
)

// Token expiry duration (3 days)
const TokenExpiryDuration = 72 // hours
