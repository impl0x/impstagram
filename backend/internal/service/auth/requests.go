package auth

import (
	"backend/internal/entity"
	"backend/pkg/dob"
)

type RegisterRequest struct {
	Username string
	Identity entity.AuthIdentity
	Dob      dob.Dob
	Password string
}

type CheckUsernameRequest struct {
	Username string
}

type LoginRequest struct {
	Identity entity.AuthIdentity // the value for the channel
	Password string
}

type ResendOTPRequest struct {
	SessionID string
}

type ForgotPasswordRequest struct {
	Identity entity.AuthIdentity
}

type ResetPasswordRequest struct {
	SessionID   string
	NewPassword string
}

type VerifyOTPRequest struct {
	SessionID string
	OTP       string
}

type RefreshRequest struct {
	RefreshToken string
}

type Add2FARequest struct {
	Channel entity.AuthChannel // only in "email" / "phone"
}

type Remove2FARequest struct {
	Channel entity.AuthChannel // only in "email" / "phone" / "totp"
}

type TotpSetupVerifyRequest struct {
	SessionID string
	OTP       string
}

type ChangePasswordRequest struct {
	CurrentPassword string
	NewPassword     string
}

type ChangeIdentifierRequest struct {
	Identity entity.AuthIdentity
}
