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
// ! Ownership and usage:
// owned by itself and used by the no one, this is an independent file.
// handler is only used by auth.go for initiation and path registration
// ! Extra
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
//   - PUT - 	/2fa
//   - DELETE - /2fa
//   - POST - 	/2fa/totp/setup
//   - POST - 	/2fa/totp/verify
func (h Handler) RegisterPaths(g *mo.Grouped) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/resend-otp", h.ResendOTP)
	g.POST("/verify-otp", h.VerifyOTP)
	g.POST("/forgot-password", h.ForgotPassword)
	g.POST("/reset-password", h.ResetPassword)
	g.POST("/refresh", h.Refresh)

	g.POST("/logout", h.Logout, h.AuthMiddleware)
	g.PUT("/2fa", h.Add2FA, h.AuthMiddleware)
	g.DELETE("/2fa", h.Remove2FA, h.AuthMiddleware)
	g.POST("/2fa/totp/setup", h.TotpSetup, h.AuthMiddleware)
	g.POST("/2fa/totp/verify", h.totpVerify, h.AuthMiddleware)
}

// ! some info:
// - we use anonymous structs to return the json response because using a map is more expensive as it allocates to the heap
// - we also will not store this structs globally instead of creating them anonymously on every handler call because of readability and decoupling such that no handler depends on another handler
// - Not all functions need to be explained individually as they all share the same pattern, only the first register function is explained in detail below

// ? ----+-----+-----Public paths-----+-----+-----
// All paths below are publicly accessible without a auth token requirement

// Registers a new user
//   - POST - models.RegisterRequest
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
	d, err := dob.Parse(req.Dob)
	if err != nil {
		switch err {
		case dob.ErrInvalidDobString:
			return errInvalidDob
		case dob.ErrImpossibleDob:
			return errImpossibleDob
		}
	}
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

// login a user
//   - POST - models.loginRequest
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
//   - POST - models.resendOTPRequest
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
//   - POST - models.verifyOTPRequest
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
//   - POST - models.RefreshRequest
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
//   - POST - models.forgotPasswordRequest
func (h Handler) ForgotPassword(c *mo.Context) error {
	var req forgotPasswordRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.forgotPassword(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			response.CodeOk,
			"An OTP has been sent to your "+result.channel.String(),
			struct {
				ReferenceID string `json:"reference_id"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.referenceID, result.expiresAt.Unix()},
		),
	)
}

// resets the password for a user
//   - POST - models.resetPasswordRequest
func (h Handler) ResetPassword(c *mo.Context) error {
	var req resetPasswordRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	err = h.Service.resetPassword(c.Request().Context(), req)
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
	token, err := mo.ContextGet[accessTokenJwt](c, keyAccessToken)
	if err != nil {
		return err
	}
	err = h.Service.logout(c.Request().Context(), token)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// adds a new 2 factor method for the user, totp not included
//   - PUT - models.add2FARequest
func (h Handler) Add2FA(c *mo.Context) error {
	var req add2FARequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	token, err := mo.ContextGet[accessTokenJwt](c, keyAccessToken)
	if err != nil {
		return err
	}
	err = h.Service.add2FA(c.Request().Context(), token, req)
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
//   - DELETE - models.remove2FARequest
func (h Handler) Remove2FA(c *mo.Context) error {
	var req remove2FARequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	token, err := mo.ContextGet[accessTokenJwt](c, keyAccessToken)
	if err != nil {
		return err
	}
	err = h.Service.remove2FA(c.Request().Context(), token, req)
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
	token, err := mo.ContextGet[accessTokenJwt](c, keyAccessToken)
	if err != nil {
		return err
	}
	result, err := h.Service.totpSetup(c.Request().Context(), token)
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
			}{result.referenceID, result.totpUri, result.expiresAt.Unix()},
		),
	)
}

// verifies a totp session and adds it to the user's 2fas
//   - POST - models.totpVerifyRequest
func (h Handler) totpVerify(c *mo.Context) error {
	token, err := mo.ContextGet[accessTokenJwt](c, keyAccessToken)
	if err != nil {
		return err
	}
	var req totpVerifyRequest
	err = c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.totpVerify(c.Request().Context(), token, req)
	if err != nil {
		if err == errTotpVerifyTOTPIncorrect {
			return c.JSON(
				err.(apperr.AppErr).ToHttp(
					struct {
						AttemptsRemaining int `json:"attempts_remaining"`
					}{result.remainingAttempts},
				),
			)
		}
		return err
	}
	return c.NoContent(http.StatusAccepted)
}
