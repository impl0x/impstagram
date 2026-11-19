package auth

import (
	"backend/internal/entity"
	"backend/internal/service"
	"backend/internal/service/auth"
	httpserver "backend/internal/transport/http"
	"backend/internal/transport/http/response"
	"backend/pkg/apperr"
	"backend/pkg/dob"
	"net/http"

	"github.com/impl0x/mo"
)

// ? INFO:
// main file containing all the http handlers
// the handler struct is the heart of this file, all functions are its methods.
// this file has its own [Handler.RegisterPaths] which registers all the paths
// into a [mo.Grouped] instance into paths and methods described in the function definition.
// it is not a necessity to use this function to register paths, but it is recommended
// to not register paths using the [Handler] http handlers which are exported but
// instead use this and modify the method/path if needed.

type Handler struct {
	Service *auth.Service
}

func NewHandler(s *auth.Service) Handler {
	return Handler{s}
}

// Registers all the auth paths to the handler's group,
// it is recommended to use this function instead of registering paths one by one yourself.
//
// Public paths:
//   - POST - /register
//   - POST - /register/verify-otp
//   - POST - /login
//   - POST - /login/verify-otp
//   - POST - /resend-otp
//   - POST - /verify-otp
//   - POST - /refresh
//   - POST - /forgot-password
//   - POST - /reset-password
//
// Authorized paths: (these paths are wrapped with the authorization [Middleware])
//   - POST - 	/logout
//   - DELETE - /account
//   - PUT - 	/2fa
//   - DELETE - /2fa
//   - POST - 	/2fa/totp/setup
//   - POST - 	/2fa/totp/verify
//   - POST - /change/password
func (h Handler) RegisterPaths(g mo.Grouped) {
	g.POST("/check/username", h.CheckUsername)
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/otp/verify/register", h.VerifyOTPRegister)
	g.POST("/2fa/otp", h.VerifyOTP2FA)
	// g.POST("/2fa/totp", h.VerifyTOTP2FA)
	g.POST("/otp/verify/forgot-password", h.VerifyOTPForForgotPassword)
	g.POST("/refresh", h.Refresh)
	g.POST("/otp/resend", h.ResendOTP)
	g.POST("/password/forgot", h.ForgotPassword)
	g.POST("/password/reset", h.ResetPassword)

	g.POST("/logout", h.Logout, h.Middleware)
	// g.GET("/account", h.GetAccount, h.Middleware)
	g.DELETE("/account", h.DeleteAccount, h.Middleware)
	g.PUT("/2fa", h.Add2FA, h.Middleware)
	g.DELETE("/2fa", h.Remove2FA, h.Middleware)
	g.POST("/2fa/totp/setup", h.TotpSetup, h.Middleware)
	g.POST("/2fa/totp/setup/verify", h.TotpSetupVerify, h.Middleware)
	g.POST("/change/password", h.ChangePassword, h.Middleware)
	g.POST("/change/email", h.ChangeEmail, h.Middleware)
	g.POST("/change/phone", h.ChangePhone, h.Middleware)
}

// ! some info:
// - we use anonymous structs to return the json response because using a map is more expensive as it allocates to the heap
// - we also will not store this structs globally instead of creating them anonymously on every handler call because of readability and decoupling such that no handler depends on another handler
// - Not all functions need to be explained individually as they all share the same pattern, only the first register function is explained in detail below

// ? ----+-----+-----Public paths-----+-----+-----
// All paths below are publicly accessible without a auth token requirement

// Checks if a username is taken or not, also validates it
//   - POST - [checkUsernameRequest]
func (h Handler) CheckUsername(c *mo.Context) error {
	var req checkUsernameRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.CheckUsername(
		httpserver.ToServiceContext(c),
		req.service(),
	)
	if err != nil {
		return err
	}
	var statusCode int
	if result.Exists {
		statusCode = http.StatusConflict
	} else {
		statusCode = http.StatusOK
	}
	return c.JSON(
		statusCode,
		struct {
			Taken bool `json:"taken"`
		}{result.Exists},
	)
}

// Registers a new user
//   - POST - [registerRequest]
func (h Handler) Register(c *mo.Context) error {
	// Binding the request json to the struct model and validating it at the same time
	var req registerRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err // returns validation error / json error, the mo error Handler / custom error handler defined knows how to handle these, at least assuming so.
	}
	var i entity.AuthIdentity
	if req.Email != "" {
		i.Channel = entity.ChannelEmail
		i.Value = req.Email
	} else if req.Phone != "" {
		i.Channel = entity.ChannelPhone
		i.Value = req.Phone
	} else {
		return errIdentifierNotProvided("Need an Email or Phone to register an account")
	}
	d := dob.AlwaysParse(req.Dob) // assuming validation has been done.
	result, err := h.Service.Register(
		httpserver.ToServiceContext(c),
		req.service(i, d),
	)
	if err != nil {
		return err
	}
	// if everything is good we return a json response with the response struct returned from Success method in response package. take a look there to see how the struct is defined.
	return c.JSON(
		http.StatusCreated,
		response.Success(
			codeNonErrRegisterSuccess,
			"Registration successful, please check your "+string(i.Channel)+" for the OTP to verify your account",
			struct { // using anon structs instead of maps to reduce allocation
				SessionID string `json:"session_id"`
				ExpiresAt int64  `json:"expires_at"`
			}{result.SessionID, result.ExpiresAt.Unix()},
		),
	)
}

// login a user
//   - POST - [loginRequest]
func (h Handler) Login(c *mo.Context) error {
	var req loginRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	var i entity.AuthIdentity
	if req.Email != "" {
		i.Channel = entity.ChannelEmail
		i.Value = req.Email
	} else if req.Phone != "" {
		i.Channel = entity.ChannelPhone
		i.Value = req.Phone
	} else if req.Username != "" {
		i.Channel = entity.ChannelUsername
		i.Value = req.Username
	} else {
		return errIdentifierNotProvided("Need an Email/Phone/Username to login into an account")
	}
	result, err := h.Service.Login(
		httpserver.ToServiceContext(c),
		req.service(i),
		httpserver.GetClientMetadata(c.Request()),
	)
	if err != nil {
		return err
	}
	if result.Requires2FA {
		return c.JSON(
			http.StatusAccepted,
			response.Success(
				codeNonErrTwoFARequired,
				"Two-factor authentication is required, please check your "+string(i.Channel),
				struct {
					SessionID string `json:"session_id"`
					ExpiresAt int64  `json:"expires_at"`
				}{result.SessionID, result.ExpiresAt.Unix()},
			),
		)
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			codeNonErrLoginSuccess,
			"Login successful",
			struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{result.AccessToken, result.RefreshToken},
		),
	)
}

// helper method for both register and login 2fa otp verification
func (h Handler) verifyOTPForAuth(c *mo.Context, flag uint8) error {
	var req verifyOTPRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	var fn func(ctx service.Context, req auth.VerifyOTPRequest, md entity.ClientMetadata) (auth.VerifyOTPAuthResult, error)
	switch flag {
	case 0:
		fn = h.Service.VerifyOTPForRegister
	case 1:
		fn = h.Service.VerifyOTPFor2FA
	}
	result, err := fn(
		httpserver.ToServiceContext(c),
		req.service(),
		httpserver.GetClientMetadata(c.Request()),
	)
	if err != nil {
		if result.RemainingAttempts != 0 { // it means this was populated, so we need to add it to the error data and send it now
			return httpserver.SendJSONWithAppErrMaybe(
				c,
				err,
				struct {
					AttemptsRemaining uint8 `json:"attempts_remaining"`
				}{result.RemainingAttempts},
			)
		}
		return err
	}
	return c.JSON(
		http.StatusAccepted,
		struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}{result.AccessToken, result.RefreshToken},
	)
}

// Verifies a otp for register
//   - POST - [verifyOTPRequest]
func (h Handler) VerifyOTPRegister(c *mo.Context) error {
	return h.verifyOTPForAuth(c, 0)
}

// Verifies a otp for login 2fa
//   - POST - [verifyOTPRequest]
func (h Handler) VerifyOTP2FA(c *mo.Context) error {
	return h.verifyOTPForAuth(c, 1)
}

// Verifies a otp for a forgot password password
//   - POST - [verifyOTPRequest]
func (h Handler) VerifyOTPForForgotPassword(c *mo.Context) error {
	var req verifyOTPRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.VerifyOTPForForgotPassword(
		httpserver.ToServiceContext(c),
		req.service(),
	)
	if err != nil {
		if result.RemainingAttempts != 0 { // it means this was populated, so we need to add it to the error data and send it now
			return httpserver.SendJSONWithAppErrMaybe(
				c,
				err,
				struct {
					AttemptsRemaining uint8 `json:"attempts_remaining"`
				}{result.RemainingAttempts},
			)
		}
		return err
	}
	return c.JSON(
		http.StatusAccepted,
		struct {
			SessionID string `json:"session_id"`
			ExpiresAt int64  `json:"expires_at"`
		}{result.SessionID, result.ExpiresAt.Unix()},
	)
}

// resend otp for a session id
//   - POST - [resendOTPRequest]
func (h Handler) ResendOTP(c *mo.Context) error {
	var req resendOTPRequest
	err := c.DecodeAndValidateBody(req)
	if err != nil {
		return err
	}
	err = h.Service.ResendOTP(
		httpserver.ToServiceContext(c),
		req.service(),
	)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// refreshes the token and provides a new set of tokens
//   - POST - [refreshRequest]
func (h Handler) Refresh(c *mo.Context) error {
	var req refreshRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.Refresh(
		httpserver.ToServiceContext(c),
		req.service(),
	)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			codeNonErrRefreshSuccess,
			"Token successfully refreshed",
			struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{result.AccessToken, result.RefreshToken},
		),
	)
}

// raises a request for resetting password
//   - POST - [forgotPasswordRequest]
func (h Handler) ForgotPassword(c *mo.Context) error {
	var req forgotPasswordRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	var i entity.AuthIdentity
	if req.Email != "" {
		i.Channel = entity.ChannelEmail
		i.Value = req.Email
	} else if req.Phone != "" {
		i.Channel = entity.ChannelPhone
		i.Value = req.Phone
	} else {
		return errIdentifierNotProvided("Need an Email/Phone to send reset password")
	}
	result, err := h.Service.ForgotPassword(
		httpserver.ToServiceContext(c),
		req.service(i),
	)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			apperr.CodeOk,
			"An OTP has been sent to your "+string(i.Channel),
			struct {
				SessionID string `json:"session_id"`
				ExpiresAt int64  `json:"expires_at"`
			}{result.SessionID, result.ExpiresAt.Unix()},
		),
	)
}

// resets the password for a user
//   - POST - [resetPasswordRequest]
func (h Handler) ResetPassword(c *mo.Context) error {
	var req resetPasswordRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	err = h.Service.ResetPassword(
		httpserver.ToServiceContext(c),
		req.service(),
	)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			apperr.CodeOk,
			"Password reset successfully",
			nil,
		),
	)
}

// ? ----+-----+-----Auth protected paths-----+-----+-----
// All the paths below are expected to be wrapped by a authorization middleware, [Middleware].

// deletes the user session
//   - POST - empty
func (h Handler) Logout(c *mo.Context) error {
	token := c.GetTypedMust[auth.AccessToken](keyAuthToken)
	err := h.Service.Logout(
		httpserver.ToServiceContext(c),
		token,
	)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// deletes the account, profile and all user sessions
//   - DELETE - empty
func (h Handler) DeleteAccount(c *mo.Context) error {
	token := c.GetTypedMust[auth.AccessToken](keyAuthToken)
	err := h.Service.DeleteAccount(
		httpserver.ToServiceContext(c),
		token,
	)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// adds a new 2 factor method for the user, totp not included
//   - PUT - [add2FARequest]
func (h Handler) Add2FA(c *mo.Context) error {
	var req add2FARequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	token := c.GetTypedMust[auth.AccessToken](keyAuthToken)
	err = h.Service.Add2FA(
		httpserver.ToServiceContext(c),
		token,
		req.service(),
	)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			apperr.CodeOk,
			"Added 2FA for this channel successfully",
			nil,
		),
	)
}

// removes an existing 2 factor method for the user
//   - DELETE - [remove2FARequest]
func (h Handler) Remove2FA(c *mo.Context) error {
	var req remove2FARequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	token := c.GetTypedMust[auth.AccessToken](keyAuthToken)
	err = h.Service.Remove2FA(
		httpserver.ToServiceContext(c),
		token,
		req.service(),
	)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			apperr.CodeOk,
			"Removed 2FA for this channel successfully",
			nil,
		),
	)
}

// starts a setup session for totp setup
//   - POST - empty
func (h Handler) TotpSetup(c *mo.Context) error {
	token := c.GetTypedMust[auth.AccessToken](keyAuthToken)
	result, err := h.Service.TotpSetup(
		httpserver.ToServiceContext(c),
		token,
	)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			apperr.CodeOk,
			"Setup initiated",
			struct {
				SessionID string `json:"session_id"`
				Uri       string `json:"uri"`
				ExpiresAt int64  `json:"expires_at"`
			}{result.SessionID, result.TotpUri, result.ExpiresAt.Unix()},
		),
	)
}

// verifies a totp session and adds it to the user's 2fas
//   - POST - [totpSetupVerifyRequest]
func (h Handler) TotpSetupVerify(c *mo.Context) error {
	token := c.GetTypedMust[auth.AccessToken](keyAuthToken)
	var req totpSetupVerifyRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.TotpSetupVerify(
		httpserver.ToServiceContext(c),
		token,
		req.service(),
	)
	if err != nil {
		if result.RemainingAttempts != 0 { // if not 0 then this field was populated and we need to add that to the error struct data.
			return httpserver.SendJSONWithAppErrMaybe(
				c,
				err,
				struct {
					AttemptsRemaining uint8 `json:"attempts_remaining"`
				}{result.RemainingAttempts},
			)
		}
		return err
	}
	return c.NoContent(http.StatusAccepted)
}

// changes password
//   - POST - [changePasswordRequest]
func (h Handler) ChangePassword(c *mo.Context) error {
	token := c.GetTypedMust[auth.AccessToken](keyAuthToken)
	var req changePasswordRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	err = h.Service.ChangePassword(httpserver.ToServiceContext(c), token, req.service())
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusAccepted)
}

func (h Handler) changeIdentifier[T interface {
	service() auth.ChangeIdentifierRequest
}](c *mo.Context, channel entity.AuthChannel) error {
	token := c.GetTypedMust[auth.AccessToken](keyAuthToken)
	var req T
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.ChangeIdentifier(httpserver.ToServiceContext(c), token, req.service())
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			apperr.CodeOk,
			"An OTP has been sent to your new "+string(channel)+", please use that to verify it",
			struct {
				SessionID string `json:"session_id"`
				ExpiresAt int64  `json:"expires_at"`
			}{result.SessionID, result.ExpiresAt.Unix()},
		),
	)
}

// changes email
//   - POST - [changeEmailRequest]
func (h Handler) ChangeEmail(c *mo.Context) error {
	return h.changeIdentifier[changeEmailRequest](c, entity.ChannelEmail)
}

// changes phone
//   - POST - [changePhoneRequest]
func (h Handler) ChangePhone(c *mo.Context) error {
	return h.changeIdentifier[changePhoneRequest](c, entity.ChannelPhone)
}
