package request

type SignUpRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum" example:"johndoe123"`
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"SecurePass123!"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required" example:"johndoe123"`
	Password string `json:"password" validate:"required,min=8" example:"SecurePass123!"`
}

type SendOTPEmailRequest struct {
	Email   string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Purpose string `json:"purpose" validate:"required,oneof=registration reset_password" example:"registration"`
}

type GoogleAuthRequest struct {
	Code string `json:"code" validate:"required" example:"4/0AY0e-g7X..."`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	OTP   string `json:"otp" validate:"required,len=6,numeric" example:"123456"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"NewSecurePass123!"`
}
