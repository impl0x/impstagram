package auth

import (
	"backend/internal/entity"
	"backend/internal/service/auth"
	"backend/pkg/dob"
)

// INFO
// Contains the json request bodies and their conversion methods to service requests.
// the struct validation tags are set according to service rules and must be validated
// as service expects all requests to be validated according to its rules.

type registerRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
	Dob      string `json:"dob" validate:"required,dob"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

func (r registerRequest) service(i entity.AuthIdentity, d dob.Dob) auth.RegisterRequest {
	return auth.RegisterRequest{
		Username: r.Username,
		Identity: i,
		Dob:      d,
		Password: r.Password,
	}
}

type checkUsernameRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
}

func (r checkUsernameRequest) service() auth.CheckUsernameRequest {
	return auth.CheckUsernameRequest{
		Username: r.Username,
	}
}

type loginRequest struct {
	Username string `json:"username" validate:"optional,min=3,max=30,username"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

func (r loginRequest) service(i entity.AuthIdentity) auth.LoginRequest {
	return auth.LoginRequest{
		Identity: i,
		Password: r.Password,
	}
}

type resendOTPRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}

func (r resendOTPRequest) service() auth.ResendOTPRequest {
	return auth.ResendOTPRequest{
		SessionID: r.SessionID,
	}
}

type forgotPasswordRequest struct {
	Email string `json:"email" validate:"optional,email"`
	Phone string `json:"phone" validate:"optional,e.164"`
}

func (r forgotPasswordRequest) service(i entity.AuthIdentity) auth.ForgotPasswordRequest {
	return auth.ForgotPasswordRequest{
		Identity: i,
	}
}

type resetPasswordRequest struct {
	SessionID   string `json:"session_id" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=20"`
}

func (r resetPasswordRequest) service() auth.ResetPasswordRequest {
	return auth.ResetPasswordRequest{
		SessionID:   r.SessionID,
		NewPassword: r.NewPassword,
	}
}

type verifyOTPRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	OTP       string `json:"otp" validate:"required,numeric"`
}

func (r verifyOTPRequest) service() auth.VerifyOTPRequest {
	return auth.VerifyOTPRequest{
		SessionID: r.SessionID,
		OTP:       r.OTP,
	}
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (r refreshRequest) service() auth.RefreshRequest {
	return auth.RefreshRequest{
		RefreshToken: r.RefreshToken,
	}
}

type add2FARequest struct {
	Channel string `json:"channel" validate:"required,oneof=email phone"`
}

func (r add2FARequest) service() auth.Add2FARequest {
	return auth.Add2FARequest{
		Channel: entity.AuthChannel(r.Channel),
	}
}

type remove2FARequest struct {
	Channel string `json:"channel" validate:"required,oneof=email phone totp"`
}

func (r remove2FARequest) service() auth.Remove2FARequest {
	return auth.Remove2FARequest{
		Channel: entity.AuthChannel(r.Channel),
	}
}

type totpSetupVerifyRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	OTP       string `json:"otp" validate:"required,numeric"`
}

func (r totpSetupVerifyRequest) service() auth.TotpSetupVerifyRequest {
	return auth.TotpSetupVerifyRequest{
		SessionID: r.SessionID,
		OTP:       r.OTP,
	}
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=20"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=20"`
}

func (r changePasswordRequest) service() auth.ChangePasswordRequest {
	return auth.ChangePasswordRequest{
		CurrentPassword: r.CurrentPassword,
		NewPassword:     r.NewPassword,
	}
}

type changeEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (r changeEmailRequest) service() auth.ChangeIdentifierRequest {
	return auth.ChangeIdentifierRequest{
		Identity: entity.AuthIdentity{
			Channel: entity.ChannelEmail,
			Value:   r.Email,
		},
	}
}

type changePhoneRequest struct {
	Phone string `json:"phone" validate:"required,e.164"`
}

func (r changePhoneRequest) service() auth.ChangeIdentifierRequest {
	return auth.ChangeIdentifierRequest{
		Identity: entity.AuthIdentity{
			Channel: entity.ChannelPhone,
			Value:   r.Phone,
		},
	}
}
