package auth

import (
	"backend/internal/entity"
	"backend/pkg/dob"
)

type RegisterRequest struct {
	Username string
	Channel  entity.AuthChannel // only in "email" / "phone"
	Value    string             // the value for the email or phone
	Dob      dob.Dob
	Password string
}

type CheckUsernameRequest struct {
	Username string
}

type LoginRequest struct {
	Channel  entity.AuthChannel // only in "email" / "phone" / "username"
	Value    string             // the value for the channel
	Password string
}

type ResendOTPRequest struct {
	Channel entity.AuthChannel // only in "email" / "phone"
	Value   string
	Purpose string
}

type ForgotPasswordRequest struct {
	Channel entity.AuthChannel
	Value   string
}

type ResetPasswordRequest struct {
	ReferenceID string
	NewPassword string
}

type VerifyOTPRequest struct {
	ReferenceID string
	OTP         string
}

type RefreshRequest struct {
	RefreshToken string
}

type Add2FARequest struct {
	Channel string
}

type Remove2FARequest struct {
	Channel string
}

type TotpVerifyRequest struct {
	ReferenceID string
	OTP         string
}
