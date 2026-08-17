package auth

import (
	"backend/internal/config"
	"backend/internal/pkg/dob"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/responses"
	"backend/internal/utils"
	"backend/internal/utils/codes"
	"net/http"
	"strconv"

	"github.com/impl0x/mo"
)

type Handler struct {
	Service *Service
}

func NewHandler(s *Service) Handler {
	return Handler{s}
}

// some info:
// - we are handling all the service level errors via the [Handler.handleError] function,
// 	 we return the function result in the handler purely due to idiom it is never actually gonna return an actual error
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
		return err // returns validation error / json error, the mo error handler defined knows how to handle these, at least assuming so.
	}
	// Call the register method in service, with request context and request data
	result, err := h.Service.register(c.Request().Context(), req)
	if err != nil {
		// we handle all errors in this function, every error is sentinel error
		return h.handleError(c, err) // always returns nil
	}
	// if everything is good we return a json response with the response struct returned from Success method in responses package. take a look there to see how the struct is defined.
	return c.JSON(
		http.StatusCreated,
		responses.Success(
			codes.RegisterSuccess,
			"Registration successful, please check your "+string(result.channel)+" for the OTP to verify your account",
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
		return h.handleError(c, err)
	}
	if result.requires2FA {
		return c.JSON(
			http.StatusAccepted,
			responses.Success(
				codes.TwoFARequired,
				"Two-factor authentication is required, please check your "+string(result.channel),
				struct {
					ReferenceID string `json:"reference_id"`
				}{result.referenceID},
			),
		)
	}
	return c.JSON(
		http.StatusOK,
		responses.Success(
			codes.LoginSuccess,
			"Login successful",
			struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{result.accessToken, result.refreshToken},
		),
	)
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
		return h.handleError(c, err)
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
		return h.handleError(c, err)
	}
	return c.JSON(
		http.StatusOK,
		responses.Success(
			codes.RefreshSuccess,
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
		return h.handleError(c, err)
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
		return h.handleError(c, err)
	}
	return c.JSON(
		http.StatusOK,
		responses.Success(
			codes.Ok,
			"An OTP has been sent to your "+string(result.channel),
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
		return h.handleError(c, err)
	}
	return c.JSON(
		http.StatusOK,
		responses.Success(
			codes.Ok,
			"Password reset successfully",
			nil,
		),
	)
}

func (h Handler) handleError(c *mo.Context, err error) (cJsonError error) { // returning error only for the idiom of mo, else it will always be nil if json doesn't throw one
	// switch over error and set these variables according to the error, then write a json body in the response and the variables using defer
	// this method reduces readability a little bit but it greatly reduces the redundancy
	var sc int        // statusCode
	var rc codes.Code // respCode
	var em string     // errorMessage
	defer func() {
		cJsonError = c.JSON(
			sc,
			responses.Error(
				rc,
				em,
			),
		)
	}()
	switch err {
	//? register errors
	case errMissingIdentifier:
		sc, rc, em = http.StatusBadRequest, codes.IdentifierMissing, "Need at least email or phone to register"
	case errAlreadyExistingUser:
		sc, rc, em = http.StatusForbidden, codes.UserAlreadyExists, "This user already exists, please try to log in"
	case errNotOldEnough:
		sc, rc, em = http.StatusForbidden, codes.UserNotOldEnough, "User not old enough, must be minimum of "+strconv.Itoa(int(config.MinAge))+" years old to use this service"
	case errTooOld:
		sc, rc, em = http.StatusForbidden, codes.UserTooOld, "User too old, cannot create account"
	case errUsernameAlreadyExists:
		sc, rc, em = http.StatusForbidden, codes.UsernameAlreadyExists, "This username is already registered, please try a different username"
	//? dob errors
	case dob.ErrImpossibleDob, dob.ErrInvalidDobString:
		sc, rc, em = http.StatusBadRequest, codes.ValidationError, err.Error()
	//? login errors
	case errMissingIdentifierLogin:
		sc, rc, em = http.StatusBadRequest, codes.IdentifierMissing, "Need at least email, phone or username to log in"
	case errUserNotFoundLogin, errIncorrectPassword:
		sc, rc, em = http.StatusUnauthorized, codes.CredentialsInvalid, "Invalid identifier or password"
	case errUserBanned:
		sc, rc, em = http.StatusForbidden, codes.UserBanned, "User is permanently banned from accessing this service"
	case errUserUnverified:
		sc, rc, em = http.StatusForbidden, codes.UserUnverified, "User account is unverified. Please verify with your account identifier, i.e. email or phone whichever you used to register"
		//? verify otp errors
	case errRefIDNotFound:
		sc, rc, em = http.StatusNotFound, codes.NotFound, "Reference ID for this verify request not found, please request a new otp again"
	case errOTPExpired:
		sc, rc, em = http.StatusForbidden, codes.OTPExpired, "The OTP provided is expired"
	case errIncorrectOTP:
		sc, rc, em = http.StatusBadRequest, codes.OTPExpired, "The OTP provided is incorrect"
	case errUserNotFound:
		sc, rc, em = http.StatusNotFound, codes.UserNotFound, "User not found"
	//? refresh errors
	case errInvalidRefreshToken:
		sc, rc, em = http.StatusBadRequest, codes.RefreshTokenInvalid, "Invalid refresh toke, please login again"
	case errExpiredRefreshToken:
		sc, rc, em = http.StatusNotAcceptable, codes.RefreshTokenExpired, "Refresh token has expired, please login again"
	//? reset errors
	case errResetSessionNotFound:
		sc, rc, em = http.StatusNotFound, codes.NotFound, "Reset session not found, please try to raise a reset password request again"
	case errResetSessionExpired:
		sc, rc, em = http.StatusForbidden, codes.ResetSessionExpired, "Reset session has expired, please try to raise a reset password request again"
	//? --x--x--x--
	default:
		println("Internal server error: " + err.Error())
		sc, rc, em = http.StatusInternalServerError, codes.InternalServerError, "An unexpected error occurred"
	}
	return
}
