package auth

import (
	"backend/internal/entity"
	"backend/pkg/jwt"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
	"uuid"
)

// ? ----+-----+----- TOKEN & JWT -----+-----+-----
// Contains access token related stuff and utility to convert to jwt acceptable format

// data in an access token
type AccessToken struct {
	UserID    uuid.UUID
	IssuedAt  time.Time
	ExpiresAt time.Time
	JwtID     uuid.UUID
}

// Generates a new access token data struct
func newAccessToken(userID uuid.UUID, jwtID uuid.UUID, expiryTime time.Duration) AccessToken {
	now := time.Now()
	return AccessToken{
		UserID:    userID,
		IssuedAt:  now,
		ExpiresAt: now.Add(expiryTime),
		JwtID:     jwtID,
	}
}

// converts the access token to claims which can be used in the [backend/pkg/jwt] jwt manager
func (at AccessToken) Claims() *jwt.Claims {
	return &jwt.Claims{
		Subject:   at.UserID.String(),
		ExpiresAt: jwt.NewNumericTime(at.ExpiresAt),
		IssuedAt:  jwt.NewNumericTime(at.IssuedAt),
		ID:        at.JwtID.String(),
	}
}

var errTokenInvalidUUID = errors.New("auth: invalid uuid")

// converts claims to access token data by parsing values
func parseClaims(c *jwt.Claims) (AccessToken, error) {
	userID, err := uuid.Parse(c.Subject)
	if err != nil {
		return AccessToken{}, errTokenInvalidUUID
	}
	jwtID, err := uuid.Parse(c.ID)
	if err != nil {
		return AccessToken{}, errTokenInvalidUUID
	}
	return AccessToken{
		UserID:    userID,
		IssuedAt:  c.IssuedAt.Time(),
		ExpiresAt: c.ExpiresAt.Time(),
		JwtID:     jwtID,
	}, nil
}

// ? ----+-----+----- Authorization -----+-----+-----

// takes a token string and verifies it and returns a [AccessTokenJwt] struct if valid token, else returns error
func (s *Service) IsAuthorized(token string) (AccessToken, error) {
	// Decode the token into a jwt claims struct
	claims, err := s.jwt.VerifyToken(token)

	// Check if jwt decode fails or if the signature is incorrect
	if err != nil {
		return AccessToken{}, errMiddlewareAuthTokenInvalid
	}

	// Converting the json struct into a usable data type for our app
	accessToken, err := parseClaims(claims)
	if err != nil { // should not trigger if server is not issuing bad tokens
		return AccessToken{}, errAuthTokenInvalidPayload
	}

	// Checking if the token has expired
	if accessToken.ExpiresAt.Before(time.Now()) {
		return AccessToken{}, errMiddlewareAuthTokenExpired
	}

	// Checking if the jwt is in jwt token block list
	_, _, ok := s.cache.jwtBlocklist.Get(accessToken.JwtID)
	if ok {
		// try not to give the client much info about *why* it is unauthorized, even though we know that this jwt is blacklisted
		return AccessToken{}, errMiddlewareAuthTokenInvalid
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
	return checkChannel(channel, entity.ChannelEmail, entity.ChannelPhone, entity.ChannelUsername)
}

func channelIsValidForOTP(channel entity.AuthChannel) bool {
	return channelIsValidForRegister(channel)
}

// converts [time.Duration] to a string format, only supports hours and minutes
// and outputs in "1 hour 20 minutes" / "20 minutes" / "1 minute" / "2 hours"
//
// panics if duration is less than a minute or more than a day
func timeDurationToString(d time.Duration) string {
	if d < time.Minute {
		panic("duration cannot be less than a minute")
	} else if d > time.Hour*24 {
		panic("duration cannot be more than a day")
	}
	var buf strings.Builder
	buf.Grow(64) // why not.
	h := d / time.Hour
	if h != 0 {
		buf.WriteString(strconv.Itoa(int(h)))
		buf.WriteString(" hour")
		if h != 1 {
			buf.WriteRune('s')
		}
	}
	if d%time.Hour != 0 {
		if h != 0 {
			buf.WriteRune(' ') // if there was hour
		}
		m := (d - h*time.Hour) / time.Minute
		buf.WriteString(strconv.Itoa(int(m)))
		buf.WriteString(" minute")
		if m != 1 {
			buf.WriteRune('s')
		}
	}
	return buf.String()
}
