package auth

import (
	"backend/internal/pkg"
	"net/http"

	"github.com/impl0x/mo"
)

type Handler struct {
	service *Service
}

func (h *Handler) SigninHandler(c *mo.Context) error {
	var req SigninRequest
	err := c.DecodeAndValidateBody(&req)
	if err != nil {
		return err // mo knows how to handle errors and return appropriate messages.
	}

	token, err := h.service.Login(req)
	if err != nil {
		return mo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(
		http.StatusOK,
		pkg.NewSuccessResponse(
			http.StatusOK, struct {
				Token string `json:"token"`
			}{token},
		),
	)
}
