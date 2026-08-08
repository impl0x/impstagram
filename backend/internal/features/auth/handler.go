package auth

import (
	"backend/internal/pkg/responses"
	"backend/internal/utils/codes"
	"net/http"

	"github.com/impl0x/mo"
)

type Handler struct {
	service *Service
}

// todo: fix register
// ? POST - models.RegisterRequest
func (h Handler) RegisterHandler(c *mo.Context) error {
	var req RegisterRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	// token, err := h.service.Register(c.Request().Context(), req)
	// if err != nil {
	// 	return responses.Error(c, http.StatusBadRequest, err.Error())
	// }
	// return responses.Success(
	// 	c,
	// 	http.StatusOK, struct {
	// 		Token string `json:"token"`
	// 	}{token},
	// )

}

// ? POST - models.LoginRequest
func (h Handler) LoginHandler(c *mo.Context) error {
	var req LoginRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	result, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		h.handleError(c, err)
	}
	if result.Requires2FA {
		return c.JSON(
			http.StatusAccepted,
			responses.Success(
				codes.TwoFARequired,
				"Two-factor authentication is required, please check your primary 2FA identifier",
				struct {
					SessionId string `json:"session_id"`
				}{result.ReferenceId},
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
			}{result.Token},
		),
	)
}

func (h Handler) handleError(c *mo.Context, err error) {
	switch err {
	case ErrMissingIdentifier:
		c.JSON(
			http.StatusBadRequest,
			responses.Error(
				codes.MissingIdentifier,
				"Need at least one of the following: email, phone or username",
			),
		)
	case ErrUserNotFound, ErrIncorrectPassword:
		c.JSON(
			http.StatusUnauthorized,
			responses.Error(
				codes.InvalidCredentials,
				"Invalid identifier or password",
			),
		)
	case ErrUserBanned:
		c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserBanned,
				"User is permanently banned from accessing this service",
			),
		)
	case ErrUserUnverified:
		c.JSON(
			http.StatusForbidden,
			responses.Error(
				codes.UserUnverified,
				"User account is unverified. Please verify with your account identifier, i.e. email or phone whichever you used to sign up",
			),
		)
	}
}
