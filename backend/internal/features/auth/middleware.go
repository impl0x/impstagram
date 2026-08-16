package auth

import (
	"backend/internal/config"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/responses"
	"backend/internal/pkg/ttlcache"
	"backend/internal/utils/codes"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/impl0x/mo"
	"github.com/impl0x/mo/validator"
)

// Do not mutate this variable at runtime.
//
// # Only add or get values from this cache, those methods are thread safe in nature
//
// This is used to store jwt ids which are blacklisted before they expire on their own.
var jwtTokenBlockList = ttlcache.New[uuid.UUID, struct{}](config.TTLCacheCleanIntervalJWTBlockList)

// ? ----+-----+-----Auth Middleware-----+-----+-----

// Checks for authorization header and expects a valid JWT, if satisfied stores it in the [mo.Context.Store] map with the key "jwt"
//
// else it returns a 401 Unauthorized error to the client if header not present, not valid jwt, jwt expired, etc other errors.
//
// Any handler wrapped with this middleware can safely assure that mo.Context["jwt"] will ALWAYS return a valid [jwt.AccessTokenPayload] struct
func Middleware(next mo.HandlerFunc) mo.HandlerFunc {
	return func(c *mo.Context) (handlerErr error) {
		// We defer a function which returns a unauthorized error with a errorMessage message variable if the variable has been set.
		// this is done to reduce redundancy of the same type of code
		var statusCode int = http.StatusUnauthorized
		var errorMessage string

		defer func() {
			if errorMessage != "" {
				handlerErr = c.JSON(
					statusCode,
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
		var accessTokenPayload jwt.AccessTokenPayload
		err := jwt.VerifyToken(authHeader, &accessTokenPayload)

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
		errs := validator.Validate(accessTokenPayload) //! we do not technically need to validate a access token if the server is correctly issuing tokens, comment this part out if everything is tested and working
		if errs != nil {
			errorMessage = "Invalid authorization token issued by the server! Validation error for the JSON payload"
			return
		}

		// Converting the json struct into a usable data type for our app
		accessTokenJwt, err := accessTokenPayload.Convert()
		if err != nil {
			errorMessage = "Invalid data in the JSON payload for authorization token"
			return
		}
		// Checking if the jwt has expired
		if accessTokenJwt.ExpiresAt.Before(time.Now()) {
			errorMessage = "Token has expired, please refresh it using the refresh token"
			return
		}

		// Checking if the jwt is in jwt token block list
		_, _, ok := jwtTokenBlockList.Get(accessTokenJwt.JwtID)
		if ok {
			errorMessage = "Unauthorized" // try not to give the client much info about *why* it is unauthorized, even though we know that this jwt is blacklisted
			return
		}

		// We store the access token data struct in the context's store map
		c.Store["jwt"] = accessTokenJwt
		return next(c)
	}
}
