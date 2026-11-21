package dto

type SignupStartRequest struct {
	Email string `json:"email" binding:"required,email,max=150"`
}

type SignupStartResponse struct {
	ResendAfterSec int `json:"resendAfterSec"`
	ExpiresInMin   int `json:"expiresInMin"`
}

type SignupVerifyRequest struct {
	Email string `json:"email" binding:"required,email,max=150"`
	Code  string `json:"code" binding:"required,len=6,number"`
}

type SignupVerifyResponse struct {
	SignupToken string `json:"signupToken"`
}

type SignupCompleteRequest struct {
	SignupToken     string `json:"signupToken" binding:"required"`
	Name            string `json:"name" binding:"required,min=2,max=100"`
	Password        string `json:"password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type AuthConfig struct {
	DemoLogin bool   `json:"demoLogin"`
	EmailMode string `json:"emailMode"`
}