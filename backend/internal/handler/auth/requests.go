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

func (r registerRequest) service(c entity.AuthChannel, v string, d dob.Dob) auth.RegisterRequest {
	return auth.RegisterRequest{
		Username: r.Username,
		Channel:  c,
		Value:    v,
		Dob:      d,
		Password: r.Password,
	}
}

type loginRequest struct {
	Username string `json:"username" validate:"optional,min=3,max=30,username"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

func (r loginRequest) service(c entity.AuthChannel, v string) auth.LoginRequest {
	return auth.LoginRequest{
		Channel:  c,
		Value:    v,
		Password: r.Password,
	}
}

type resendOTPRequest struct {
	Purpose  string `json:"purpose" validate:"required,oneof=registration 2fa reset_password"`
	Username string `json:"username" validate:"optional,min=3,max=30,username"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
}

func (r resendOTPRequest) service(c entity.AuthChannel, v string) auth.ResendOTPRequest {
	return auth.ResendOTPRequest{
		Channel: c,
		Value:   v,
		Purpose: r.Purpose,
	}
}

type forgotPasswordRequest struct {
	Email string `json:"email" validate:"optional,email"`
	Phone string `json:"phone" validate:"optional,e.164"`
}

func (r forgotPasswordRequest) service(c entity.AuthChannel, v string) auth.ForgotPasswordRequest {
	return auth.ForgotPasswordRequest{
		Channel: c,
		Value:   v,
	}
}

type resetPasswordRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,startswith=pwd_"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=20"`
}

func (r resetPasswordRequest) service() auth.ResetPasswordRequest {
	return auth.ResetPasswordRequest{
		ReferenceID: r.ReferenceID,
		NewPassword: r.NewPassword,
	}
}

type verifyOTPRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,startswith=otp_"`
	OTP         string `json:"otp" validate:"required,numeric,len=6"`
}

func (r verifyOTPRequest) service() auth.VerifyOTPRequest {
	return auth.VerifyOTPRequest{
		ReferenceID: r.ReferenceID,
		OTP:         r.OTP,
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
		Channel: r.Channel,
	}
}

type remove2FARequest struct {
	Channel string `json:"channel" validate:"required,oneof=email phone totp"`
}

func (r remove2FARequest) service() auth.Remove2FARequest {
	return auth.Remove2FARequest{
		Channel: r.Channel,
	}
}

type totpVerifyRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,startswith=totp_"`
	OTP         string `json:"otp" validate:"required,numeric,len=6"`
}

func (r totpVerifyRequest) service() auth.TotpVerifyRequest {
	return auth.TotpVerifyRequest{
		ReferenceID: r.ReferenceID,
		OTP:         r.OTP,
	}
}
