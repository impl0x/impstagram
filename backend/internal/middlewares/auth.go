package middlewares

import (
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/responses"
	"backend/internal/utils/codes"
	"net/http"
	"time"

	"github.com/impl0x/mo"
	"github.com/impl0x/mo/validator"
)

func Authorization(next mo.HandlerFunc) mo.HandlerFunc {
	return func(c *mo.Context) (handlerErr error) {
		// We defer a function which returns a unauthorized error with a errorMessage message variable if the variable has been set.
		// this is done to reduce redundancy of the same type of code
		var errorMessage string

		defer func() {
			if errorMessage != "" {
				handlerErr = c.JSON(
					http.StatusUnauthorized,
					responses.Error(
						codes.Unauthorized,
						errorMessage,
					),
				)
			}
		}()

		// get the auth header value from the request header
		authHeader := c.Request().Header.Get("authorization")
		if authHeader == "" {
			errorMessage = "Authorization header missing or empty"
			return
		}

		// Decode the token into a jwt payload struct
		var accessTokenData jwt.AccessTokenPayload
		err := jwt.VerifyToken(authHeader, &accessTokenData)

		// Check if jwt decode fails or if the signature is incorrect
		if err != nil {
			switch err {
			case jwt.ErrInvalidJWTToken:
				errorMessage = "Invalid authorization token"
			case jwt.ErrIncorrectJWTToken:
				errorMessage = "Authorization token has been tampered with, signature mismatch"
			}
			return
		}

		// validating the jwt payload (optional)
		errs := validator.Validate(accessTokenData) //! we do not technically need to validate a access token if the server is correctly issuing tokens, comment this part out if everything is tested and working
		if errs != nil {
			errorMessage = "Invalid authorization token issued by the server! Validation error for the json payload"
			panic("authorization mw: invalid jwt token issued by the server!, validation error for the json payload")
		}

		// Checking if the jwt has expired
		expiresAt := time.Unix(int64(accessTokenData.ExpiresAt), 0)
		if expiresAt.Before(time.Now()) {
			errorMessage = "Token has expired, please refresh it using the refresh token"
			return
		}

		// We store the access token data struct in the context's store map
		c.Store["jwt"] = accessTokenData
		return next(c)
	}
}
