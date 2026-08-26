package auth

import (
	"backend/internal/pkg/apperr"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/response"
	"backend/internal/pkg/ttlcache"
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
	return func(c *mo.Context) (handlerErr error) {
		// get the auth header value from the request header
		authHeader := c.Request().Header.Get("authorization")
		if authHeader == "" {
			return apperr.NewUnauthorized(response.CodeUnauthorized, "Authorization header missing or empty")
		}

		// Decode the token into a jwt payload struct
		var accessTokenPayload accessTokenJwtPayload
		err := jwt.VerifyToken(authHeader, &accessTokenPayload)

		// Check if jwt decode fails or if the signature is incorrect
		if err != nil {
			switch err { // assuming jwt.VerifyToken can only return these 2 errors
			case jwt.ErrInvalidJWTToken:
				return apperr.NewUnauthorized(response.CodeUnauthorized, "Invalid authorization token")
			case jwt.ErrIncorrectJWTToken:
				return apperr.NewUnauthorized(response.CodeUnauthorized, "Authorization token has been tampered with, signature mismatch")
			}
		}

		// validating the jwt payload
		// ! (optional)
		errs := validator.Validate(accessTokenPayload) // ! we do not technically need to validate a access token if the server is correctly issuing tokens, comment this part out if everything is tested and working
		if errs != nil {
			return apperr.NewUnauthorized(response.CodeUnauthorized, "Invalid authorization token issued by the server! Validation error for the JSON payload")
		}

		// Converting the json struct into a usable data type for our app
		accessTokenJwt, err := accessTokenPayload.Convert()
		if err != nil {
			return apperr.NewUnauthorized(response.CodeUnauthorized, "Invalid data in the JSON payload for authorization token")
		}
		// Checking if the jwt has expired
		if accessTokenJwt.expiresAt.Before(time.Now()) {
			return apperr.NewUnauthorized(response.CodeUnauthorized, "Token has expired, please refresh it using the refresh token")
		}

		// Checking if the jwt is in jwt token block list
		_, _, ok := jwtTokenBlockList.Get(accessTokenJwt.jwtID)
		if ok {
			// try not to give the client much info about *why* it is unauthorized, even though we know that this jwt is blacklisted
			return apperr.NewUnauthorized(response.CodeUnauthorized, "Invalid authorization token")
		}

		// We store the access token data struct in the context's store map
		c.Store["jwt"] = accessTokenJwt
		return next(c)
	}
}
