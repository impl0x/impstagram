package auth

import (
	"backend/internal/entity"
	"backend/pkg/jwt"
	"errors"
	"slices"
	"time"
	"uuid"

	"github.com/impl0x/go-utils/cache"
	"github.com/impl0x/mo/validator/v3"
)

// Do not mutate this variable at runtime.
//
// # Only add or get values from this cache, those methods are thread safe in nature
//
// This is used to store jwt ids which are blacklisted before they expire on their own.
var jwtTokenBlockList = cache.NewTTLCache[uuid.UUID, struct{}](ruleTTLCacheCleanIntervalJWTBlockList)

// ? ----+-----+-----JWT & AUTH-----+-----+-----
// jwt helper functions for generating access tokens
// auth funcs for checking authorization

// The access token jwt payload used in the actual token data after encoding
type AccessTokenJwtPayload struct {
	UserID    string `json:"sub" validate:"required,len=36"`
	IssuedAt  uint64 `json:"iat" validate:"required"`
	ExpiresAt uint64 `json:"exp" validate:"required"`
	JwtID     string `json:"jti" validate:"required,len=36"`
}

// Usable struct for the service with converted data types
type AccessTokenJwt struct {
	UserID    uuid.UUID
	IssuedAt  time.Time
	ExpiresAt time.Time
	JwtID     uuid.UUID
}

// Generates a new access token struct
func newAccessToken(userID uuid.UUID, jwtID uuid.UUID, expiryTime time.Duration) AccessTokenJwt {
	now := time.Now()
	return AccessTokenJwt{
		UserID:    userID,
		IssuedAt:  now,
		ExpiresAt: now.Add(expiryTime),
		JwtID:     jwtID,
	}
}

var errJwtInvalidUUID = errors.New("auth.models: invalid uuid")

// Converts a AccessTokenPayload to a more usable AccessToken type with the values being converted to usable uuid.UUID and time.Time
//
// only returns error of errJwtInvalidUUID if the UUID parsing fails
func (atp AccessTokenJwtPayload) Convert() (AccessTokenJwt, error) {
	userID, err := uuid.Parse(atp.UserID)
	if err != nil {
		return AccessTokenJwt{}, errJwtInvalidUUID
	}
	jwtID, err := uuid.Parse(atp.JwtID)
	if err != nil {
		return AccessTokenJwt{}, errJwtInvalidUUID
	}
	return AccessTokenJwt{
		UserID:    userID,
		IssuedAt:  time.Unix(int64(atp.IssuedAt), 0),
		ExpiresAt: time.Unix(int64(atp.ExpiresAt), 0),
		JwtID:     jwtID,
	}, nil
}

// converts the access token to the payload version which can be used in json marshalling
func (at AccessTokenJwt) Payload() AccessTokenJwtPayload {
	return AccessTokenJwtPayload{
		UserID:    at.UserID.String(),
		IssuedAt:  uint64(at.IssuedAt.Unix()),
		ExpiresAt: uint64(at.ExpiresAt.Unix()),
		JwtID:     at.JwtID.String(),
	}
}

// takes a token string and verifies it and returns a [AccessTokenJwt] struct if valid token, else returns error
func IsAuthorized(token string) (AccessTokenJwt, error) {
	// Decode the token into a jwt payload struct
	var accessTokenPayload AccessTokenJwtPayload
	err := jwt.VerifyToken(token, &accessTokenPayload)

	// Check if jwt decode fails or if the signature is incorrect
	if err != nil {
		return AccessTokenJwt{}, errMiddlewareAuthTokenInvalid
	}

	// validating the jwt payload
	// ! (optional)
	errs := validator.Validate(accessTokenPayload) // ! we do not technically need to validate a access token if the server is correctly issuing tokens, comment this part out if everything is tested and working
	if errs != nil {
		return AccessTokenJwt{}, errAuthTokenInvalidPayload // this is an internal error
	}

	// Converting the json struct into a usable data type for our app
	accessToken, err := accessTokenPayload.Convert()
	if err != nil {
		return AccessTokenJwt{}, errAuthTokenInvalidPayload
	}
	// Checking if the token has expired
	if accessToken.ExpiresAt.Before(time.Now()) {
		return AccessTokenJwt{}, errMiddlewareAuthTokenExpired
	}

	// Checking if the jwt is in jwt token block list
	_, _, ok := jwtTokenBlockList.Get(accessToken.JwtID)
	if ok {
		// try not to give the client much info about *why* it is unauthorized, even though we know that this jwt is blacklisted
		return AccessTokenJwt{}, errMiddlewareAuthTokenInvalid
	}
	return accessToken, nil
}

// ? ----+-----+-----Utility funcs-----+-----+-----

// checks if a auth channel is in the slice of auth channels.
func checkChannel(channel entity.AuthChannel, chans ...entity.AuthChannel) bool {
	return slices.Contains(chans, channel)
}

func channelIsValidForRegister(channel entity.AuthChannel) bool {
	return checkChannel(channel, entity.ChannelEmail, entity.ChannelPhone)
}

func channelIsValidForLogin(channel entity.AuthChannel) bool {
	return checkChannel(channel, entity.ChannelEmail, entity.ChannelPhone, entity.ChannelTOTP)
}

func channelIsValidForOTP(channel entity.AuthChannel) bool {
	return channelIsValidForRegister(channel)
}
