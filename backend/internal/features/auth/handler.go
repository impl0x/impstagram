package auth

import (
	"backend/internal/pkg/common"
	"backend/internal/pkg/responses"
	"errors"
	"net/http"

	"github.com/impl0x/mo"
)

type Handler struct {
	service *Service
}
// todo: fix the handlers to follow a different error and translation model
// ? POST - models.RegisterRequest
func (h *Handler) RegisterHandler(c *mo.Context) error {
	var req RegisterRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	token, err := h.service.Register(c.Request().Context(), req)
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, err.Error())
	}
	return responses.Success(
		c,
		http.StatusOK, struct {
			Token string `json:"token"`
		}{token},
	)

}

// ? POST - models.LoginRequest
func (h *Handler) LoginHandler(c *mo.Context) error {
	var req LoginRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err
	}
	token, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, common.ErrInternalServerError){
			return mo.ErrInternalServerError
		}
		return responses.Error(c, http.StatusBadRequest, err.Error())
	}
	return responses.Success(
		c,
		http.StatusOK, struct {
			Token string `json:"token"`
		}{token},
	)

}


