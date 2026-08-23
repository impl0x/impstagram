package auth

import (
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/response"
	"backend/internal/utils"
	"net/http"

	"github.com/impl0x/mo"
)

type Handler struct {
	Service *Service
}

func NewHandler(s *Service) Handler {
	return Handler{s}
}

// some info:
// - we are handling all the Service level errors via the [Handler.handleError] function,
// 	 we return the function result in the Handler purely due to idiom it is never actually gonna return an actual error
// 	 except if json encode error occurs which is not probable because its our function with valid syntax and logic.
//
// - we use anonymous structs to return the json response because using a map is more expensive as it allocates to the heap
// - Not all functions need to be explained individually as they all share the same pattern, only the first register function is explained in detail below

// ? POST - models.RegisterRequest
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
			}{result.referenceID},
		),
	)
}

// ? POST - models.loginRequest
func (h Handler) Login(c *mo.Context) error {
	if h.Service == nil {
		println(h.Service)
		return c.NoContent(204)
	}
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
				}{result.referenceID},
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

// ? POST - models.resendOTPRequest
func (h Handler) ResendOTP(c *mo.Context) error{
	var req resendOTPRequest
	err:=c.DecodeAndValidateBody(req)
	if err!=nil{
		return err
	}
	h.Service.resendOTP(c.Request().Context(), req)
}

// ? POST - models.verifyOTPRequest
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
		return err
	}
	if result.isResetRequest {
		return c.JSON(
			http.StatusAccepted,
			struct {
				ReferenceID string `json:"reference_id"`
			}{result.referenceID},
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

// ? POST - models.RefreshRequest
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

// ? [AUTH PROTECTED] POST - empty
func (h Handler) Logout(c *mo.Context) error {
	accessTokenJwt := c.Store["jwt"].(jwt.AccessToken) // type conversion and reading the map assumes that this path has the authorization [Middleware] wrapped beforehand and it is working
	err := h.Service.logout(c.Request().Context(), accessTokenJwt)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// ? POST - models.forgotPasswordRequest
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
			}{result.referenceID},
		),
	)
}

// ? POST - models.resetPasswordRequest
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
