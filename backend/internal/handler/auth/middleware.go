package auth

import (
	"backend/internal/service/auth"
	"backend/pkg/apperr"
	"backend/pkg/response"
	"strings"

	"github.com/impl0x/mo"
)

// ? ----+-----+-----Auth Middleware-----+-----+-----

// Checks for authorization header and expects a valid JWT, if satisfied stores it in the [mo.Context.Store] map with the key [keyAuthToken]
//
// else it returns a 401 Unauthorized error to the client if header not present, not valid jwt, jwt expired, etc other errors.
func (h Handler) AuthMiddleware(next mo.HandlerFunc) mo.HandlerFunc {
	errHeaderMissing := apperr.NewUnauthorized(response.CodeUnauthorized, "Authorization header missing or empty")
	errUnsupportedAuth := apperr.NewUnauthorized(response.CodeUnauthorized, "Authorization type is unsupported or not provided")
	return func(c *mo.Context) error {
		token := c.Request().Header.Get("authorization")
		if token == "" {
			return errHeaderMissing
		}
		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return errUnsupportedAuth
		}
		jwt, err := auth.IsAuthorized(token)
		if err != nil {
			return err
		}
		c.Add(keyAuthToken, jwt)
		return next(c)
	}
}
