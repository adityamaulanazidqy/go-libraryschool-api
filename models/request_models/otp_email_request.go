package request_models

type RequestOtpEmail struct {
	Email string `json:"email" binding:"required,email"`
}
type VerificationOtpEmail struct {
	Email string `json:"email" binding:"required,email"`
	Otp   string `json:"otp" binding:"required"`
}
