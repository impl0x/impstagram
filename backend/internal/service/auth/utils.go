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

	"github.com/impl0x/go-utils/cache"
)

// ? ----+-----+----- TTL Cache Leased Item -----+-----+-----

// wrapper over the cache leased item to make it easier to use
type ttlLeasedItemWrapper[K comparable, V any] struct {
	cache.TTLLeasedItem[K, V]
	flag bool
}

const _defaultFlag = true

func newTTLLeasedItemWrapper[K comparable, V any](i cache.TTLLeasedItem[K, V]) *ttlLeasedItemWrapper[K, V] {
	return &ttlLeasedItemWrapper[K, V]{
		i, _defaultFlag,
	}
}

// defer this
//
// checks if release flag is true and releases the item after its used,
// else does not release letting it be deleted automatically if delete was
// called previously making the flag false
func (t *ttlLeasedItemWrapper[K, V]) d() {
	if t.flag {
		t.TTLLeasedItem.Release()
	}
}

// deletes the item not letting it stay
func (t *ttlLeasedItemWrapper[K, V]) delete() {
	t.flag = false
}

// keeps the item
func (t *ttlLeasedItemWrapper[K, V]) keep() {
	t.flag = true
}

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
	claims, err := s.token.jwt.VerifyToken(token)

	// Check if jwt decode fails or if the signature is incorrect
	if err != nil {
		return AccessToken{}, errAuthorizationAuthTokenInvalid
	}

	// Converting the json struct into a usable data type for our app
	accessToken, err := parseClaims(claims)
	if err != nil { // should not trigger if server is not issuing bad tokens
		return AccessToken{}, errAuthTokenInvalidPayload
	}

	// Checking if the token has expired
	if accessToken.ExpiresAt.Before(time.Now()) {
		return AccessToken{}, errAuthorizationAuthTokenExpired
	}

	// Checking if the jwt is in jwt token block list
	_, _, ok := s.cache.jwtBlocklist.Get(accessToken.JwtID)
	if ok {
		// try not to give the client much info about *why* it is unauthorized, even though we know that this jwt is blacklisted
		return AccessToken{}, errAuthorizationAuthTokenInvalid
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

// converts [time.Duration] to a string format, from years to nano seconds
// and outputs in "1 hour 20 minutes" / "20 minutes" / "1 minute" / "2 hours"
func timeDurationToString(d time.Duration) string {
	helper := func(b *strings.Builder, td time.Duration, t time.Duration, n string) time.Duration {
		v := td / t
		if v > 0 {
			td = td - v*t
			b.WriteString(strconv.Itoa(int(v)))
			b.WriteByte(' ')
			b.WriteString(n)
			if v > 1 {
				b.WriteByte('s')
			}
			b.WriteByte(' ')
		}
		return td
	}

	var b strings.Builder
	b.Grow(32) // pretty enough for most strings

	d = helper(&b, d, 365*24*time.Hour, "year")

	d = helper(&b, d, 30*24*time.Hour, "month")

	d = helper(&b, d, 24*time.Hour, "day")

	d = helper(&b, d, time.Hour, "hour")

	d = helper(&b, d, time.Minute, "minute")

	d = helper(&b, d, time.Second, "second")

	d = helper(&b, d, time.Millisecond, "millisecond")

	d = helper(&b, d, time.Microsecond, "microsecond")

	d = helper(&b, d, time.Nanosecond, "nanosecond")

	return b.String()[:b.Len()-1]
}
