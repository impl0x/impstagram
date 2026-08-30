package auth

import (
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/ttlcache"
	"errors"
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
var jwtTokenBlockList = ttlcache.New[uuid.UUID, struct{}](ruleTTLCacheCleanIntervalJWTBlockList)

// ? ----+-----+-----Auth Middleware-----+-----+-----

// Checks for authorization header and expects a valid JWT, if satisfied stores it in the [mo.Context.Store] map with the key "jwt"
//
// else it returns a 401 Unauthorized error to the client if header not present, not valid jwt, jwt expired, etc other errors.
//
// Any handler wrapped with this middleware can safely assure that mo.Context["jwt"] will ALWAYS return a valid [jwt.AccessTokenPayload] struct
func Middleware(next mo.HandlerFunc) mo.HandlerFunc {
	errAuthTokenInvalidPayload:=errors.New("auth.Middleware: Auth token has invalid payload")
	return func(c *mo.Context) (handlerErr error) {
		// get the auth header value from the request header
		authHeader := c.Request().Header.Get("authorization")
		if authHeader == "" {
			return errMiddlewareAuthHeaderMissing
		}

		// Decode the token into a jwt payload struct
		var accessTokenPayload accessTokenJwtPayload
		err := jwt.VerifyToken(authHeader, &accessTokenPayload)

		// Check if jwt decode fails or if the signature is incorrect
		if err != nil {
			return errMiddlewareAuthTokenInvalid
		}

		// validating the jwt payload
		// ! (optional)
		errs := validator.Validate(accessTokenPayload) // ! we do not technically need to validate a access token if the server is correctly issuing tokens, comment this part out if everything is tested and working
		if errs != nil {
			return errAuthTokenInvalidPayload // this is an internal error
		}

		// Converting the json struct into a usable data type for our app
		accessToken, err := accessTokenPayload.Convert()
		if err != nil {
			return errAuthTokenInvalidPayload
		}
		// Checking if the token has expired
		if accessToken.expiresAt.Before(time.Now()) {
			return errMiddlewareAuthTokenExpired
		}

		// Checking if the jwt is in jwt token block list
		_, _, ok := jwtTokenBlockList.Get(accessToken.jwtID)
		if ok {
			// try not to give the client much info about *why* it is unauthorized, even though we know that this jwt is blacklisted
			return errMiddlewareAuthTokenInvalid
		}

		// We store the access token data struct in the context's store map
		c.Add(keyAccessToken, accessToken)
		return next(c)
	}
}
