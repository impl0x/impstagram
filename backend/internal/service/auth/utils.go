package auth

import (
	"backend/internal/entity"
	"backend/internal/service"
	"backend/pkg/cryptoutil"
	"backend/pkg/email/render"
	"backend/pkg/jwt"
	"crypto/subtle"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/impl0x/go-utils/cache"
	"github.com/impl0x/go-utils/logger"
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

func channelIsValidForRegister(channel entity.AuthChannel) bool {
	switch channel {
	case entity.ChannelEmail, entity.ChannelPhone:
		return true
	}
	return false
}

func channelIsValidForLogin(channel entity.AuthChannel) bool {
	switch channel {
	case entity.ChannelEmail, entity.ChannelPhone, entity.ChannelUsername:
		return true
	}
	return false
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

	helper(&b, d, time.Nanosecond, "nanosecond")

	return b.String()[:b.Len()-1]
}

// ? ----+-----+----- OTP Management -----+-----+-----

// used to store otp session details
type otpSession struct {
	otp      string
	identity entity.AuthIdentity
	userID   uuid.UUID
	purpose  authPurpose
	attempts uint8
	resends  uint8
}

// Manager represents a OTP manager which can instantiate sessions,
// store them in cache memory, send otps and verify them using the stored sessions.
//
// can be used by value
type otpManager struct {
	otp                cryptoutil.OTPManager
	sessionIDGenerator func() string
	channels           otpManagerChannels // channels to send otps
	sessions           *cache.TTLLeaseCache[string, *otpSession]
	expiryTime         time.Duration
	expiryTimeString   string // cached version
	maxAttempts        uint8
	maxResends         uint8
	log                logger.Logger
}

type otpManagerChannels struct {
	email *_email // reference to [Service]'s email field
}

// sends otp according to channel provided, and prepares message according to purpose
//
// note: this just sends an otp, and does not generate or save a session
//
// error is returned if sending fails
func (o *otpManager) sendOTP(ctx service.Context, otp string, i entity.AuthIdentity, purpose authPurpose) error {
	o.log.Debug("Service: Sending OTP", "code", otp, "req", ctx.RequestId)
	var err error
	// Send otp based on the identifier
	switch i.Channel {
	case entity.ChannelEmail:
		var sub render.Subject
		var html string
		r := o.channels.email.renderer.Otp(otp, o.expiryTimeString) // start rendering a otp email template
		switch purpose {                                            // pick the subject and render html based on the purpose we have received
		case purpose2FA:
			sub = render.Sub2FA
			html = r.TwoFA()
		case purposeRegistration:
			sub = render.SubVerifyEmail
			html = r.Registration()
		case purposeForgotPassword:
			sub = render.SubResetPassword
			html = r.ResetPass()
		case purposeChangeIdentifier:
			sub = render.SubChangeEmail
			html = r.ChangeEmail()
		}
		err = o.channels.email.sender.SendSingle(ctx, o.channels.email.renderer.Subject(sub), i.Value, html)
	case entity.ChannelPhone:
		// IMPORTANT
		// we can never send otp to phone numbers as its not possible as of now to afford sms service or telegram's gateway service.
		// will not change the code but anyone signing up on the backend with a phone will simply not receive an otp and will not be able to continue.
		// not returning an error or anything, its intentional, due to future compatibility. the frontend should not have phone supported.
		return fmt.Errorf("Service: Request to send otp on phone arrived, impossible! to do")
	default:
		return errOTPCannotSendOTPHere
	}
	if err != nil { // send otp error
		return err
	}
	return nil
}

// saves session
func (m *otpManager) save(otp string, i entity.AuthIdentity, userID uuid.UUID, purpose authPurpose) (string, time.Time) {
	// save session to cache
	s := m.sessionIDGenerator()
	e := time.Now().Add(m.expiryTime)
	m.sessions.Add(
		s,
		&otpSession{
			otp:      otp,
			identity: i,
			userID:   userID,
			purpose:  purpose,
		},
		e,
	)
	return s, e
}

// Sends a generated otp to the auth identity and saves the session.
//
// this is a wrapper around both [otpManager.sendOTP] and [otpManager.save]
//
// returns the session id, expiry time and optional error
//
// error is returned if [otpManager.sendOTP] returns one
func (m *otpManager) send(ctx service.Context, i entity.AuthIdentity, userID uuid.UUID, purpose authPurpose) (string, time.Time, error) {
	otp := m.otp.Generate()
	err := m.sendOTP(ctx, otp, i, purpose)
	if err != nil {
		return "", time.Time{}, err
	}
	s, e := m.save(otp, i, userID, purpose)
	return s, e, nil
}

// Resends an otp using the session id.
//
// Sends the same otp, and expiration still remains the same.
// after max amount of resends has happened, it does not allow
// for any more resends.
func (m *otpManager) resend(ctx service.Context, sessionID string) error {
	item, ok := m.sessions.Acquire(sessionID)
	if !ok {
		return errOTPSessionNotFound
	}
	defer item.Release()
	session, expiresAt := item.Value(), item.ExpiresAt()
	switch session.identity.Channel {
	case entity.ChannelTOTP, entity.ChannelUsername:
		return errOTPCannotSendOTPHere
	}
	if expiresAt.Before(time.Now()) {
		item.Delete()
		return errOTPExpired
	}
	// not allow for any more resends if limit has been reached
	if session.resends >= m.maxResends {
		return errOTPResendAttemptsExhausted
	}
	session.resends++
	err := m.sendOTP(ctx, session.otp, session.identity, session.purpose)
	if err != nil {
		return err
	}
	return nil
}

// Verifies a otp using the session id provided.
func (m *otpManager) verify(sessionID, otp string, purpose authPurpose) (*otpSession, uint8, error) {
	item, ok := m.sessions.Acquire(sessionID)
	if !ok {
		return nil, 0, errOTPSessionNotFound
	}
	defer item.Release() // defer releasing the item, (delete calls dont let this happen so its safe)

	session, expiresAt := item.Value(), item.ExpiresAt()
	if expiresAt.Before(time.Now()) {
		// we delete the session if its expired
		item.Delete()
		return session, 0, errOTPExpired
	}

	remainingAttempts := m.maxAttempts - session.attempts
	if remainingAttempts == 0 {
		// this prevents any more verification attempts
		// on this session essentially destroying this session.
		item.Delete()
		return session, 0, errCommonAttemptsExhausted
	}

	// validation function
	if session.purpose != purpose {
		return session, 0, errOTPIncorrectPurpose
	}

	session.attempts++

	if subtle.ConstantTimeCompare([]byte(session.otp), []byte(otp)) == 1 {
		return session, remainingAttempts, errOTPIncorrect
	}
	// if otp matches then we delete session
	item.Delete()
	// return result without any errors
	return session, 0, nil
}
