package request

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GoogleAuthRequest struct {
	Code string `json:"code"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}
