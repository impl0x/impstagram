package auth

import (
	"backend/internal/config"
	"backend/internal/pkg/dob"
	"backend/internal/pkg/responses"
	"backend/internal/utils/codes"
	"net/http"
	"strconv"

	"github.com/impl0x/mo"
)

type Handler struct {
	Service *Service
}

// some info:
// - we are handling all the service level errors via the [Handler.handleError] function,
// 	 we return the function result in the handler purely due to idiom it is never actually gonna return an actual error
// 	 except if json encode error occurs which is not probable because its our function with valid syntax and logic.
//
// - we use anonymous structs to return the json response because using a map is more expensive as it allocates to the heap

// ? POST - models.RegisterRequest
func (h Handler) Register(c *mo.Context) error {
	var req registerRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.register(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err) // always returns nil
	}
	return c.JSON(
		http.StatusAccepted,
		responses.Success(
			codes.RegisterSuccess,
			"Registration successful, please check your "+string(result.channel)+" for the OTP to verify your account",
			struct {
				ReferenceID string `json:"reference_id"`
			}{result.referenceID},
		),
	)
}

// ? POST - models.LoginRequest
func (h Handler) Login(c *mo.Context) error {
	var req loginRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.Service.login(c.Request().Context(), req)
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
				Token string `json:"token"`
			}{result.token},
		),
	)
}

func (h Handler) Verify(c *mo.Context) error {
	var req verifyOTPRequest
	err := c.DecodeAndValidateBody(req)
	if err != nil {
		return err
	}
	result, err := h.Service.verifyOTP(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(
		http.StatusOK,
		struct {
			Token string `json:"token"`
		}{result.token},
	)
}

func (h Handler) handleError(c *mo.Context, err error) error { // returning error only for the idiom of mo, else it will always be mo if json doesn't throw one
	switch err {
	case errMissingIdentifier:
		return c.JSON(
			http.StatusBadRequest,
			responses.Error(
				codes.MissingIdentifier,
				"Need at least one of the following: email, phone or username",
			),
		)
	case errUserNotFound, errIncorrectPassword:
		return c.JSON(
			http.StatusUnauthorized,
			responses.Error(
				codes.InvalidCredentials,
				"Invalid identifier or password",
			),
		)
	case errUserBanned:
		return c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserBanned,
				"User is permanently banned from accessing this service",
			),
		)
	case errUserUnverified:
		return c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserUnverified,
				"User account is unverified. Please verify with your account identifier, i.e. email or phone whichever you used to sign up",
			),
		)
	case dob.ErrImpossibleDob, dob.ErrInvalidDobString:
		return c.JSON(
			http.StatusBadRequest,
			responses.Error(
				codes.ValidationError,
				err.Error(),
			),
		)
	case errAlreadyExistingUser:
		return c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserExists,
				"This user already exists, please try to log in",
			),
		)
	case errNotOldEnough:
		return c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserNotOldEnough,
				"User not old enough, must be minimum of "+strconv.Itoa(int(config.MinAge))+" years old to use this service",
			),
		)
	case errTooOld:
		return c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserTooOld,
				"User too old, cannot create account",
			),
		)
	case errUsernameAlreadyExists:
		return c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UsernameAlreadyExists,
				"This username is already registered, please try a different username",
			),
		)
	case errRefIDNotFound:
		return c.JSON(
			http.StatusNotFound,
			responses.Error(
				codes.ReferenceIDNotFound,
				"Reference ID for this verify request not found, please try to request a new otp again",
			),
		)
	case errOTPExpired:
		return c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.OTPExpired,
				"The OTP provided is expired",
			),
		)
	case errIncorrectOTP:
		return c.JSON(
			http.StatusBadRequest,
			responses.Error(
				codes.OTPIncorrect,
				"The OTP provided is incorrect",
			),
		)
	default:
		println("Internal server error: " + err.Error())
		return c.JSON(
			http.StatusInternalServerError,
			responses.Error(
				codes.InternalServerError,
				"An unexpected error occurred",
			),
		)

	}
}
