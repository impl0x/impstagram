package auth

import (
	"backend/internal/config"
	"backend/internal/pkg/dob"
	"backend/internal/pkg/responses"
	"backend/internal/utils/codes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/impl0x/mo"
)

type Handler struct {
	service *Service
}

// ? POST - models.RegisterRequest
func (h Handler) Register(c *mo.Context) error {
	var req registerRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.service.Register(c.Request().Context(), req)
	if err != nil {
		h.handleError(c, err)
		return nil
	}
	return c.JSON(
		http.StatusAccepted,
		responses.Success(
			codes.RegisterSuccess,
			"Registration successful, please check your "+result.twoFAIdentifier.string()+" for the OTP to verify your account",
			struct {
				ReferenceId string `json:"reference_id"`
			}{result.referenceId},
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
	result, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		h.handleError(c, err)
		return nil
	}
	if result.requires2FA {
		return c.JSON(
			http.StatusAccepted,
			responses.Success(
				codes.TwoFARequired,
				"Two-factor authentication is required, please check your "+result.twoFAIdentifier.string(),
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

func (h Handler) Verify (c *mo.Context) error{
	var req verifyRequest
	err:=c.DecodeAndValidateBody(req)
	if err!=nil{
		return err
	}
	result,err:=h.service.verify(c.Request().Context(),req)
	// todo: handle error and return response appropriately
}

func (h Handler) handleError(c *mo.Context, err error) {
	switch err {
	case errMissingIdentifier:
		c.JSON(
			http.StatusBadRequest,
			responses.Error(
				codes.MissingIdentifier,
				"Need at least one of the following: email, phone or username",
			),
		)
	case errUserNotFound, errIncorrectPassword:
		c.JSON(
			http.StatusUnauthorized,
			responses.Error(
				codes.InvalidCredentials,
				"Invalid identifier or password",
			),
		)
	case errUserBanned:
		c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserBanned,
				"User is permanently banned from accessing this service",
			),
		)
	case errUserUnverified:
		c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserUnverified,
				"User account is unverified. Please verify with your account identifier, i.e. email or phone whichever you used to sign up",
			),
		)
	case dob.ErrImpossibleDob, dob.ErrInvalidDobString:
		c.JSON(
			http.StatusBadRequest,
			responses.Error(
				codes.ValidationError,
				err.Error(),
			),
		)
	case errAlreadyExistingUser:
		c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserExists,
				"This user already exists, please try to log in",
			),
		)
	case errNotOldEnough:
		c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserNotOldEnough,
				"User not old enough, must be minimum of "+strconv.Itoa(int(config.MinAge))+" years old to use this service",
			),
		)
	case errTooOld:
		c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserTooOld,
				"User too old, cannot create account",
			),
		)
	default:
		fmt.Println("Internal server error: %v", err)
		c.JSON(
			http.StatusInternalServerError,
			responses.Error(
				codes.InternalServerError,
				"An unexpected error occurred",
			),
		)

	}
}
