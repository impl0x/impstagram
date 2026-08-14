package auth

import (
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/responses"
	"backend/internal/utils/codes"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/impl0x/mo"
	"github.com/impl0x/mo/validator"
)

// ? ----+-----+-----Token block list-----+-----+-----

type accessTokenBlockList struct {
	blockRequestList map[uuid.UUID]struct{}
	mu               sync.RWMutex
}

func (bl *accessTokenBlockList) Add(jwtID uuid.UUID) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.blockRequestList[jwtID] = struct{}{}
}

func (bl *accessTokenBlockList) IsPresent(jwtID uuid.UUID) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	_, ok := bl.blockRequestList[jwtID]
	return ok
}

func (bl *accessTokenBlockList) Delete(jwtID uuid.UUID) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	delete(bl.blockRequestList, jwtID)
}

// This variable is a constant and must not be changed at runtime
//
// This is used to block JWT access tokens before they expire, used especially if a user logs out and then we put the jwt id in this list
//
// Any jwt id in this block list will be blocked and be sent a 401 unauthorized
var AccessTokenBlockList = accessTokenBlockList{
	blockRequestList: make(map[uuid.UUID]struct{}),
}

// ? ----+-----+-----Cleaning solutions for token block list-----+-----+-----

func passiveCleaner()

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
