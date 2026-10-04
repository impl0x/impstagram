package middleware

import (
	"backend/internal/service/auth"
	"backend/pkg/apperr"
	"backend/pkg/response"
	"strings"

	"github.com/impl0x/mo"
)

const KeyAuthToken = "at"

// Checks for authorization header and expects a valid JWT, 
// if satisfied stores it in the [mo.Context] with the key [KeyAuthToken], 
// use [mo.Context.Get] or [mo.Context.GetTyped] to retrieve using the same key.
//
// calls the service method [auth.IsAuthorized] to check the token validity
func Authorization(next mo.HandlerFunc) mo.HandlerFunc {
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
		c.Add(KeyAuthToken, jwt)
		return next(c)
	}
}
