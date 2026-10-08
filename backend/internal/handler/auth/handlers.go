package auth

import (
	"backend/internal/entity"
	"backend/internal/service/auth"
	"backend/internal/util"
	"backend/pkg/apperr"
	"backend/pkg/dob"
	"backend/pkg/response"
	"backend/pkg/useragent"
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
//   - POST - /login
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
func (h Handler) RegisterPaths(g mo.Grouped) {
	g.POST("/register", h.Register)
	g.POST("/check/username", h.CheckUsername)
	g.POST("/login", h.Login)
	g.POST("/resend-otp", h.ResendOTP)
	g.POST("/verify-otp", h.VerifyOTP)
	g.POST("/refresh", h.Refresh)
	g.POST("/forgot-password", h.ForgotPassword)
	g.POST("/reset-password", h.ResetPassword)

	g.POST("/logout", h.Logout, h.Middleware)
	g.DELETE("/account", h.DeleteAccount, h.Middleware)
	g.PUT("/2fa", h.Add2FA, h.Middleware)
	g.DELETE("/2fa", h.Remove2FA, h.Middleware)
	g.POST("/2fa/totp/setup", h.TotpSetup, h.Middleware)
	g.POST("/2fa/totp/verify", h.totpVerify, h.Middleware)
}

// ! some info:
// - we use anonymous structs to return the json response because using a map is more expensive as it allocates to the heap
// - we also will not store this structs globally instead of creating them anonymously on every handler call because of readability and decoupling such that no handler depends on another handler
// - Not all functions need to be explained individually as they all share the same pattern, only the first register function is explained in detail below

// ? ----+-----+-----Public paths-----+-----+-----
// All paths below are publicly accessible without a auth token requirement

// Registers a new user
//   - POST - [registerRequest]
func (h Handler) Register(c *mo.Context) error {
	// Binding the request json to the struct model and validating it at the same time
	var req registerRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err // returns validation error / json error, the mo error Handler / custom error handler defined knows how to handle these, at least assuming so.
	}
	var ch entity.AuthChannel
	var v string
	if req.Email != "" {
		ch = entity.ChannelEmail
		v = req.Email
	} else if req.Phone != "" {
		ch = entity.ChannelPhone
		v = req.Phone
	} else {
		return errIdentifierNotProvided("Need an Email or Phone to register an account")
	}
	d := dob.MustParse(req.Dob) // assuming validator has already validated this.
	result, err := h.Service.Register(c.Request().Context(), req.service(ch, v, d))
	if err != nil {
		return err
	}
	// if everything is good we return a json response with the response struct returned from Success method in response package. take a look there to see how the struct is defined.
	return c.JSON(
		http.StatusCreated,
		response.Success(
			codeNonErrRegisterSuccess,
			"Registration successful, please check your "+string(ch)+" for the OTP to verify your account",
			struct { // using anon structs instead of maps to reduce allocation
				ReferenceID string `json:"reference_id"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.ReferenceID, result.ExpiresAt.Unix()},
		),
	)
}

func (h Handler) CheckUsername(c *mo.Context) error {
	var req checkUsernameRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.CheckUsername(c.Request().Context(), req.service())
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

// login a user
//   - POST - [loginRequest]
func (h Handler) Login(c *mo.Context) error {
	var req loginRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	var ch entity.AuthChannel
	var v string
	if req.Email != "" {
		ch = entity.ChannelEmail
		v = req.Email
	} else if req.Phone != "" {
		ch = entity.ChannelPhone
		v = req.Phone
	} else if req.Username != "" {
		ch = entity.ChannelUsername
		v = req.Username
	} else {
		return errIdentifierNotProvided("Need an Email/Phone/Username to login into an account")
	}
	result, err := h.Service.Login(
		c.Request().Context(),
		req.service(ch, v),
		entity.ClientMetadata{
			IPAddress: util.GetIpFromRequest(c.Request()), // assumes the function has the correct implementation of ip retrieval depending upon environment and reverse proxy configurations
			UserAgent: useragent.Parse(c.Request().UserAgent()),
		},
	)
	if err != nil {
		return err
	}
	if result.Requires2FA {
		return c.JSON(
			http.StatusAccepted,
			response.Success(
				codeNonErrTwoFARequired,
				"Two-factor authentication is required, please check your "+string(ch),
				struct {
					ReferenceID string `json:"reference_id"`
					ExpiresAt   int64  `json:"expires_at"`
				}{result.ReferenceID, result.ExpiresAt.Unix()},
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

// resend otp for any purpose
//   - POST - [resendOTPRequest]
func (h Handler) ResendOTP(c *mo.Context) error {
	var req resendOTPRequest
	err := c.DecodeAndValidateBody(req)
	if err != nil {
		return err
	}
	var ch entity.AuthChannel
	var v string
	if req.Email != "" {
		ch = entity.ChannelEmail
		v = req.Email
	} else if req.Phone != "" {
		ch = entity.ChannelPhone
		v = req.Phone
	} else if req.Username != "" {
		ch = entity.ChannelUsername
		v = req.Username
	} else {
		return errIdentifierNotProvided("Need an Email/Phone/Username to send otp")
	}
	result, err := h.Service.ResendOTP(c.Request().Context(), req.service(ch, v))
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			codeNonErrOTPSent,
			"OTP sent successfully to user's "+string(ch),
			struct {
				ReferenceID string `json:"reference_id"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.ReferenceID, result.ExpiresAt.Unix()},
		),
	)
}

// verifies the otp for any purpose
//   - POST - [verifyOTPRequest]
func (h Handler) VerifyOTP(c *mo.Context) error {
	var req verifyOTPRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.VerifyOTP(
		c.Request().Context(),
		req.service(),
		entity.ClientMetadata{
			IPAddress: util.GetIpFromRequest(c.Request()), // assumes the function has the correct implementation of ip retrieval depending upon environment and reverse proxy configurations
			UserAgent: useragent.Parse(c.Request().UserAgent()),
		},
	)
	if err != nil {
		if result.RemainingAttempts != 0 { // it means this field was populated so we need to add that field in the error struct.
			return c.JSON(
				err.(apperr.AppErr).ToHttp( // service always returns a AppErr
					struct {
						AttemptsRemaining int `json:"attempts_remaining"`
					}{result.RemainingAttempts},
				),
			)
		}
		return err
	}
	if result.IsResetRequest {
		return c.JSON(
			http.StatusAccepted,
			struct {
				ReferenceID string `json:"reference_id"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.ReferenceID, result.ExpiresAt.Unix()},
		)
	}
	return c.JSON(
		http.StatusOK,
		struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}{result.AccessToken, result.RefreshToken},
	)
}

// refreshes the token and provides a new set of tokens
//   - POST - [refreshRequest]
func (h Handler) Refresh(c *mo.Context) error {
	var req refreshRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.Refresh(c.Request().Context(), req.service())
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
	var ch entity.AuthChannel
	var v string
	if req.Email != "" {
		ch = entity.ChannelEmail
		v = req.Email
	} else if req.Phone != "" {
		ch = entity.ChannelPhone
		v = req.Phone
	} else {
		return errIdentifierNotProvided("Need an Email/Phone to send reset password")
	}
	result, err := h.Service.ForgotPassword(c.Request().Context(), req.service(ch, v))
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			response.CodeOk,
			"An OTP has been sent to your "+string(ch),
			struct {
				ReferenceID string `json:"reference_id"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.ReferenceID, result.ExpiresAt.Unix()},
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
	err = h.Service.ResetPassword(c.Request().Context(), req.service())
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			response.CodeOk,
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
	token, err := c.GetTyped[auth.AccessToken](keyAuthToken)
	if err != nil {
		return err
	}
	err = h.Service.Logout(c.Request().Context(), token)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// deletes the account, profile and all user sessions
//   - DELETE - empty
func (h Handler) DeleteAccount(c *mo.Context) error {
	token, err := c.GetTyped[auth.AccessToken](keyAuthToken)
	if err != nil {
		return err
	}
	err = h.Service.DeleteAccount(c.Request().Context(), token)
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
	token, err := c.GetTyped[auth.AccessToken](keyAuthToken)
	if err != nil {
		return err
	}
	err = h.Service.Add2FA(c.Request().Context(), token, req.service())
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			response.CodeOk,
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
	token, err := c.GetTyped[auth.AccessToken](keyAuthToken)
	if err != nil {
		return err
	}
	err = h.Service.Remove2FA(c.Request().Context(), token, req.service())
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			response.CodeOk,
			"Removed 2FA for this channel successfully",
			nil,
		),
	)
}

// starts a setup session for totp setup
//   - POST - empty
func (h Handler) TotpSetup(c *mo.Context) error {
	token, err := c.GetTyped[auth.AccessToken](keyAuthToken)
	if err != nil {
		return err
	}
	result, err := h.Service.TotpSetup(c.Request().Context(), token)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			response.CodeOk,
			"Setup initiated",
			struct {
				ReferenceId string `json:"reference_id"`
				Uri         string `json:"uri"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.ReferenceID, result.TotpUri, result.ExpiresAt.Unix()},
		),
	)
}

// verifies a totp session and adds it to the user's 2fas
//   - POST - [totpVerifyRequest]
func (h Handler) totpVerify(c *mo.Context) error {
	token, err := c.GetTyped[auth.AccessToken](keyAuthToken)
	if err != nil {
		return err
	}
	var req totpVerifyRequest
	err = c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.TotpVerify(c.Request().Context(), token, req.service())
	if err != nil {
		if result.RemainingAttempts != 0 { // if not 0 then this field was populated and we need to add that to the error struct data.
			return c.JSON(
				err.(apperr.AppErr).ToHttp(
					struct {
						AttemptsRemaining int `json:"attempts_remaining"`
					}{result.RemainingAttempts},
				),
			)
		}
		return err
	}
	return c.NoContent(http.StatusAccepted)
}
