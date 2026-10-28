package auth

import (
	httpserver "backend/internal/transport/http"
	"backend/pkg/apperr"
	"backend/pkg/response"
	"strings"

	"github.com/impl0x/mo"
)

const keyAuthToken httpserver.ContextKey = "at"

// Checks for authorization header and expects a valid JWT,
// if satisfied stores it in the [mo.Context] with the key [KeyAuthToken],
// use [mo.Context.Get] or [mo.Context.GetTyped] to retrieve using the same key.
//
// calls the service method [auth.IsAuthorized] to check the token validity
func (h Handler) Middleware(next mo.HandlerFunc) mo.HandlerFunc {
	errHeaderMissing := apperr.NewUnauthorized(response.CodeUnauthorized, "Authorization header missing or empty")
	errUnsupportedAuth := apperr.NewUnauthorized(response.CodeUnauthorized, "Authorization type is unsupported or not provided")
	return func(c *mo.Context) error {
		// get the auth header
		token := c.Request().Header.Get("authorization")
		if token == "" {
			return errHeaderMissing
		}

		// split the bearer part out and also check if its present
		idx := strings.IndexByte(token, ' ')
		if idx == -1 || token[:idx] != "Bearer" {
			return errUnsupportedAuth
		}

		// call service method to check authorization with the token part
		jwt, err := h.Service.IsAuthorized(token[idx+1:])
		if err != nil {
			return err
		}

		// if okay then add the token to the context store
		c.Add(keyAuthToken, jwt)
		return next(c)
	}
}
