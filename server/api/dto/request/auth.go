package request

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SendOTPEmailRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

type GoogleAuthRequest struct {
	Code string `json:"code"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
