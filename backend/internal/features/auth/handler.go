package auth

import (
	"backend/internal/pkg/apperr"
	"backend/internal/pkg/response"
	"backend/internal/utils"
	"net/http"

	"github.com/impl0x/mo"
)

type Handler struct {
	Service *Service
	Group   *mo.Grouped
}

func NewHandler(s *Service, g *mo.Grouped) Handler {
	return Handler{s, g}
}

// Registers all the auth paths to the handler's group
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
//   - PUT - 	/2fa ? adds a new 2fa
//   - DELETE - /2fa ? removes a 2fa
//   - POST - 	/2fa/totp/setup
//   - POST - 	/2fa/totp/verify
func (h Handler) RegisterPaths() {
	h.Group.POST("/register", h.Register)
	h.Group.POST("/login", h.Login)
	h.Group.POST("/resend-otp", h.ResendOTP)
	h.Group.POST("/verify-otp", h.VerifyOTP)
	h.Group.POST("/refresh", h.Refresh)
	h.Group.POST("/forgot-password", h.ForgotPassword)
	h.Group.POST("/reset-password", h.ResetPassword)

	h.Group.POST("/logout", h.Logout, Middleware)
	h.Group.POST("/2fa/add", h.Add2FA, Middleware)
	h.Group.POST("/2fa/totp/setup", h.TotpSetup, Middleware)
	h.Group.POST("/2fa/totp/verify", h.totpVerify, Middleware)
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
		return err // returns validation error / json error, the mo error Handler defined knows how to handle these, at least assuming so.
	}
	// Call the register method in Service, with request context and request data
	result, err := h.Service.register(c.Request().Context(), req)
	if err != nil {
		// we handle all errors in this function, every error is sentinel error
		return err // always returns nil
	}
	// if everything is good we return a json response with the response struct returned from Success method in response package. take a look there to see how the struct is defined.
	return c.JSON(
		http.StatusCreated,
		response.Success(
			codeRegisterSuccess,
			"Registration successful, please check your "+result.channel.String()+" for the OTP to verify your account",
			struct { // using anon structs instead of maps to reduce allocation
				ReferenceID string `json:"reference_id"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.referenceID, result.expiresAt.Unix()},
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
	result, err := h.Service.login(
		c.Request().Context(),
		req,
		requestMetadata{
			utils.GetIpFromRequest(c.Request()), // assumes the function has the correct implementation of ip retrieval depending upon environment and reverse proxy configurations
			c.Request().UserAgent(),
		},
	)
	if err != nil {
		return err
	}
	if result.requires2FA {
		return c.JSON(
			http.StatusAccepted,
			response.Success(
				codeTwoFARequired,
				"Two-factor authentication is required, please check your "+result.channel.String(),
				struct {
					ReferenceID string `json:"reference_id"`
					ExpiresAt   int64  `json:"expires_at"`
				}{result.referenceID, result.expiresAt.Unix()},
			),
		)
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			codeLoginSuccess,
			"Login successful",
			struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{result.accessToken, result.refreshToken},
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
	result, err := h.Service.resendOTP(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			response.CodeOk,
			"OTP sent successfully to user's "+result.channel.String(),
			struct {
				ReferenceID string `json:"reference_id"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.referenceID, result.expiresAt.Unix()},
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
	result, err := h.Service.verifyOTP(
		c.Request().Context(),
		req,
		requestMetadata{
			utils.GetIpFromRequest(c.Request()), // assumes the function has the correct implementation of ip retrieval depending upon environment and reverse proxy configurations
			c.Request().UserAgent(),
		},
	)
	if err != nil {
		if err == errIncorrectOTP {
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
	if result.isResetRequest {
		return c.JSON(
			http.StatusAccepted,
			struct {
				ReferenceID string `json:"reference_id"`
				ExpiresAt   int64  `json:"expires_at"`
			}{result.referenceID, result.expiresAt.Unix()},
		)
	}
	return c.JSON(
		http.StatusOK,
		struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}{result.accessToken, result.refreshToken},
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
	result, err := h.Service.refresh(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		response.Success(
			codeRefreshSuccess,
			"Token successfully refreshed",
			struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{result.accessToken, result.refreshToken},
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
		if err == errTotpIncorrectOTP {
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
