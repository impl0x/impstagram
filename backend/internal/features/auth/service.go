package auth

import (
	"backend/internal/config"
	"backend/internal/pkg/dob"
	"backend/internal/pkg/email"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/password"
	"backend/internal/pkg/token"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	Email email.Client
	Repo  Repository

	mu          sync.RWMutex          // mutex for the OTPSessions
	OTPSessions map[string]otpSession // string is the reference id
}

// ? [HELPER] ----+-----+-----OTP-----+-----+-----

// Used to store the pending otp sessions
type otpSession struct {
	userID    uuid.UUID
	channel   authChannel // e.g., channelEmail
	purpose   authPurpose
	secretOTP string
	expiresAt time.Time
}

// The reason these are all separate functions is to ensure mutexes are used and its separated from the business logic,
// it is a bit redundant to do this but I feel its better than putting mutexes in the business logic every time

// Adds a new OTP session to the map with the reference id passed, with mutex.
func (s *Service) addOTPSession(refID string, channel authChannel, purpose authPurpose, userID uuid.UUID, otp string) {
	session := otpSession{
		userID,
		channel,
		purpose,
		otp,
		time.Now().Add(config.OTPExpiryTime), // calculating expires at by adding OTP expiry time defined in config + current time
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OTPSessions[refID] = session
}

// Gets a OTP session from the map
func (s *Service) getOTPSession(refID string) (session otpSession, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok = s.OTPSessions[refID]
	return
}

// Removes a OTP session from the map, with mutex.
func (s *Service) removeOTPSession(refID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.OTPSessions, refID)
}

// ? ----+-----+-----Register-----+-----+-----

// Stores the result to the register call
type registerResult struct {
	referenceID string      // Used to link the upcoming OTP request
	channel     authChannel // the identifier/channel used to register, eg: email/phone
}

// Registration sentinel errors
var (
	errMissingIdentifier     = errors.New("missing identifier")
	errAlreadyExistingUser   = errors.New("user already exists")
	errNotOldEnough          = errors.New("user not old enough")
	errTooOld                = errors.New("user too old")
	errUsernameAlreadyExists = errors.New("username already exists")
)

// Registers a new user
func (s *Service) register(ctx context.Context, req registerRequest) (registerResult, error) {
	// validate the date of birth string (this technically should be in the handler layer but then I would have to pass dob object separately in this function which would look bad so i just did it here)
	userDob, err := dob.NewDobFromString(req.Dob)
	if err != nil {
		return registerResult{}, err // returns dob sentinel errors, which we handle in the handler layer
	}

	// validate the user's age against our business rules
	userAge := userDob.Age()
	if userAge < config.MinAge {
		return registerResult{}, errNotOldEnough
	} else if userAge > config.MaxAge {
		return registerResult{}, errTooOld
	}

	// Figure out what identifier the user sent, that is if the user registered with a email or a phone
	var user *user
	var channel authChannel // identifier, eg: email/phone
	var target string       // the identifier literal, eg: jon@email.com/+1-234567890
	if req.Email != "" {
		channel = channelEmail
		target = req.Email
	} else if req.Phone != "" {
		channel = channelPhone
		target = req.Phone
	} else {
		return registerResult{}, errMissingIdentifier // if user sent neither a email or a phone then we reject the registration request
	}

	// Finding in database
	user, err = s.Repo.FindUserByChannel(ctx, channel, target)

	if err != nil {
		return registerResult{}, err // ignoring database level errors as of now.
	}

	if user != nil { // if db returned a user
		return registerResult{}, errAlreadyExistingUser // yes i acknowledge that user has chances of being banned/unverified, but this is intended. We want user to login and then hit those errors if they exist.
	}

	// checks if the username already exists because usernames are unique
	sameUsernameUser, err := s.Repo.FindUserByChannel(ctx, channelUsername, req.Username) // assuming req.Username is validated in validator
	if sameUsernameUser != nil || err == nil {                                            // todo: handle db level errors
		return registerResult{}, errUsernameAlreadyExists
	}

	// Now we hash the user's password and create a new user in the database
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return registerResult{}, err
	}
	user = NewUser(req, passwordHash)  // returns a unverified user by default
	err = s.Repo.CreateUser(ctx, user) // we create a user before sending otp
	if err != nil {
		return registerResult{}, err
	}

	// Verify the identifier. By sending an otp
	otp, err := s.sendOTP(channel, purposeRegistration, target)
	if err != nil {
		return registerResult{}, err
	}
	// generate a new reference id for a otp session and add it to our otp sessions map
	refID := token.GenerateReferenceID()
	s.addOTPSession(refID, channel, purposeRegistration, user.ID, otp)

	return registerResult{
		referenceID: refID,
		channel:     channel,
	}, nil
}

// Explanation on some of the design choices in the above function:
// one of the design choices in this function was to first create the user in the database and then send a verification email/sms
// this is because if the db call fails then we exit early with no otp sent,
// and if the db call succeeds but sending email fails we have a unverified user but no otp
// if the frontend shows the internal error the user may try to register again which will trigger a already exists
// which then will prompt the user to login and if they login they will be prompted to verify their email which is a secure workflow,
// although more work for the user but this is considering that this is the worst case scenario.
// better than a dangling otp with no registered user.

// ? ----+-----+-----Login-----+-----+-----

// Stores the result to the login call
type loginResult struct {
	accessToken  string
	refreshToken string
	requires2FA  bool // if this is false then below all fields are zero'd out, else the above token is zero value'd
	channel      authChannel
	referenceID  string // Used to link the upcoming OTP request
}

// Login sentinel errors
var (
	errUserNotFound      = errors.New("user not found")
	errIncorrectPassword = errors.New("incorrect password")
	errUserBanned        = errors.New("user banned")
	errUserUnverified    = errors.New("user unverified")
)

// Login a user in, and optionally if the user has 2fa enabled it asks for a code on the verify endpoint.
//
// rmd requestMetadata is required for userSession storage on successful login
func (s *Service) login(ctx context.Context, req loginRequest, rmd requestMetadata) (loginResult, error) {
	// Figure out what identifier the user sent, i.e. email/phone/username to log in
	var user *user
	var err error
	switch {
	case req.Username != "":
		user, err = s.Repo.FindUserByChannel(ctx, channelUsername, req.Username)
	case req.Email != "":
		user, err = s.Repo.FindUserByChannel(ctx, channelEmail, req.Email)
	case req.Phone != "":
		user, err = s.Repo.FindUserByChannel(ctx, channelPhone, req.Phone)
	default:
		return loginResult{}, errMissingIdentifier
	}

	if err != nil { // database error
		return loginResult{}, err
	}
	// todo : fix db error and user not found error
	if user == nil {
		return loginResult{}, errUserNotFound
	}

	// if user is found in database we compare the passwords and the password hash in the database to see if the user has the correct password
	ok, err := password.Compare(req.Password, user.PasswordHash)
	if err != nil {
		return loginResult{}, err
	}
	if !ok {
		return loginResult{}, errIncorrectPassword
	}
	// check if user is banned then we return immediately not allowing a login, else if unverified, in which case we trigger a resend otp request
	switch user.Status {
	case statusBanned:
		return loginResult{}, errUserBanned
	case statusUnverified:
		return loginResult{}, errUserUnverified // todo: make a way for the user be able to resend a otp, otherwise this ends up being a edge deadlock condition
	}

	// If user has 2Fa enabled ask for otp.
	if user.TwoFAs != nil {
		referenceID := token.GenerateReferenceID()
		primaryTwoFAChannel := user.TwoFAs[0] // by default the first element is the primary 2FA identifier
		// if it is TOTP
		if primaryTwoFAChannel == channelTOTP { // if its a time based otp we don't bother generating or sending it anywhere
			s.addOTPSession(referenceID, channelTOTP, purpose2FA, user.ID, user.TotpSecretKey) // we don't save an otp but instead save the secret key from the database to reduce a db call on the verify endpoint
			return loginResult{
				requires2FA: true,
				channel:     channelTOTP,
				referenceID: referenceID,
			}, nil
		}
		// else if its email or phone
		var target string // either the email or the phone literal
		switch primaryTwoFAChannel {
		case channelEmail:
			target = user.Email
		case channelPhone:
			target = user.Phone
		default:
			panic("invalid identifier found in 2fa slice of user")
		}
		otp, err := s.sendOTP(primaryTwoFAChannel, purpose2FA, target)
		if err != nil {
			return loginResult{}, err
		}
		s.addOTPSession(referenceID, primaryTwoFAChannel, purpose2FA, user.ID, otp)
		return loginResult{
			requires2FA: true,
			channel:     primaryTwoFAChannel,
			referenceID: referenceID,
		}, nil
	}
	// else

	// Generate tokens
	jwtID := uuid.New()
	accessToken, err := jwt.GenerateToken(jwt.NewAccessTokenPayload(user.ID, jwtID))
	if err != nil {
		panic(err)
	}
	refreshToken := token.GenerateRefreshToken()

	// Add a new user session to the database
	err = s.Repo.CreateSession(
		ctx,
		newUserSession(
			jwtID,                               // jwtID
			token.GenerateMD5Hash(refreshToken), // tokenHash
			rmd.IP,                              // userIP
			rmd.userAgent,                       // userAgent - assumes that past middleware has checked if a valid ua is present
			user.ID,                             // userID
		),
	)
	if err != nil {
		return loginResult{}, err // db error
	}

	return loginResult{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

// ? [HELPER] ----+-----+-----Send otp-----+-----+-----

// SendOTP sends a one-time password challenge to a user destination.
// The behavior of this function changes based on the following configurations:
//   - channel: The transport medium used to deliver the token (Email or SMS)
//   - purpose: The system context (2FA, Reset password or Registration) used to select templates
//   - target: The absolute address string (e.g., an email address or E.164 phone number)
func (s *Service) sendOTP(channel authChannel, purpose authPurpose, target string) (string, error) {
	otp, err := token.GenerateOTP()
	if err != nil { // generate otp error
		return "", err
	}
	// Send otp based on the identifier
	switch channel {
	case channelEmail:
		var emailSendRequest email.SendRequest
		switch purpose {
		case purpose2FA:
			emailSendRequest = email.NewSendRequest(target, email.SubjectTwoFa, email.Html2FAOTP.Format(otp))
		case purposeRegistration:
			emailSendRequest = email.NewSendRequest(target, email.SubjectVerifyEmail, email.HtmlRegistrationVerificationOTP.Format(otp))
		case purposeResetPass:
			emailSendRequest = email.NewSendRequest(target, email.SubjectResetPassword, email.HtmlResetPasswordOTP.Format(otp))
		}
		err = s.Email.Send(emailSendRequest)
	case channelPhone:
		// todo: send otp on phone
	default:
		panic("invalid otp channel")
	}
	if err != nil { // send otp error
		return "", err
	}
	return otp, nil
}

// ? ----+-----+-----Verify otp-----+-----+-----

type verifyResult struct {
	accessToken  string
	refreshToken string
}

var (
	errRefIDNotFound = errors.New("reference ID not found")
	errOTPExpired    = errors.New("otp expired")
	errIncorrectOTP  = errors.New("incorrect otp")
)

// Verifies the two factor / verification OTP and generates a token for the user.
func (s *Service) verifyOTP(ctx context.Context, req verifyOTPRequest, rmd requestMetadata) (verifyResult, error) { // token, error
	// retrieve the session from the reference id in the request
	session, ok := s.getOTPSession(req.ReferenceID)
	if !ok { // if not found it means either the frontend is trying to reuse the same reference id after expiration or verification
		return verifyResult{}, errRefIDNotFound
	}
	if session.expiresAt.Before(time.Now()) { // if the otp has expired we remove it from our map and return a error
		s.removeOTPSession(req.ReferenceID)
		return verifyResult{}, errOTPExpired
	}
	var err error
	if session.channel == channelTOTP { // if its a authenticator time based 2 factor code
		session.secretOTP, err = token.GenerateTOTP(session.secretOTP) // in this case secretOTP is actually the secretKey for the TOTP which is used to compute the otp.
		if err != nil {
			return verifyResult{}, err
		}
	}
	if session.secretOTP != req.OTP { // we check the otp if it does not match we return early with a incorrect otp error
		return verifyResult{}, errIncorrectOTP
	}
	// if the otp matches then we remove the reference id from our map immediately
	s.removeOTPSession(req.ReferenceID)
	// find the user
	user, err := s.Repo.FindUserByID(ctx, session.userID)
	if err != nil {
		return verifyResult{}, err
	}
	if user == nil { // if the user mysteriously got deleted after just trying to log in or register...
		return verifyResult{}, errUserNotFound
	}

	// If the verification purpose was registration we also need to set the user status to verified in the database
	if session.purpose == purposeRegistration {
		user.Status = statusVerified
		err = s.Repo.UpdateUser(ctx, user.ID, user)
		if err != nil {
			return verifyResult{}, err
		}
	}

	// generate new tokens and return them to the user for future usage
	jwtID := uuid.New()
	accessToken, err := jwt.GenerateToken(jwt.NewAccessTokenPayload(user.ID, jwtID))
	if err != nil {
		panic(err)
	}
	refreshToken := token.GenerateRefreshToken()

	// Add a new user session to the database
	err = s.Repo.CreateSession(
		ctx,
		newUserSession(
			jwtID,                               // jwtID
			token.GenerateMD5Hash(refreshToken), // tokenHash
			rmd.IP,                              // userIP
			rmd.userAgent,                       // userAgent - assumes that past middleware has checked if a valid ua is present
			user.ID,                             // userID
		),
	)

	if err != nil {
		return verifyResult{}, err
	}

	return verifyResult{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

// ? ----+-----+-----Refresh-----+-----+-----

type refreshResult struct {
	accessToken  string
	refreshToken string
}

var (
	errInvalidRefreshToken = errors.New("invalid refresh token")
	errExpiredRefreshToken = errors.New("expired refresh token")
)

func (s *Service) refresh(ctx context.Context, req refreshRequest) (refreshResult, error) {
	// Do a db lookup with the refresh token's hash
	userSesh, err := s.Repo.FindSessionByToken(ctx, token.GenerateMD5Hash(req.RefreshToken))
	if err != nil {
		return refreshResult{}, err // db error
	}
	if userSesh == nil {
		return refreshResult{}, errInvalidRefreshToken
	}

	// Check if the session has expired
	if userSesh.ExpiresAt.Before(time.Now()) {
		// Deleting the user session if it is expired, the user will have to create a new session again by logging in.
		err = s.Repo.DeleteSession(ctx, userSesh.ID)
		if err != nil {
			return refreshResult{}, err // db error
		}
		return refreshResult{}, errExpiredRefreshToken
	}

	// if everything is good we generate both new tokens, we do not have to regenerate and re update the jwt id as its unnecessary. it is a fixed value which is linked with the user session in database
	accessToken, err := jwt.GenerateToken(jwt.NewAccessTokenPayload(userSesh.UserID, userSesh.JwtID))
	if err != nil {
		panic(err)
	}
	refreshToken := token.GenerateRefreshToken()

	// update the session with the new refresh token and also update the expires at field to the max capacity again.
	err = s.Repo.UpdateSessionToken(
		ctx,
		userSesh.ID,
		token.GenerateMD5Hash(refreshToken), // we store a hash of the token
		time.Now().AddDate(0, 0, config.RefreshTokenExpiryTime),
	)
	if err != nil {
		return refreshResult{}, err // db err
	}

	// return both tokens
	return refreshResult{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

// ? ----+-----+-----Logout-----+-----+-----

func (s *Service) logout(ctx context.Context, accessTokenJwt jwt.AccessToken) error {
	err := s.Repo.DeleteSessionByJwtID(ctx, accessTokenJwt.JwtID)
	if err != nil {
		return err // db error
	}
	AccessTokenBlockList.Add(accessTokenJwt.JwtID, accessTokenJwt.ExpiresAt)
	if AccessTokenBlockListCleaner == TimerCleanerType {
		TimerCleaner(accessTokenJwt.JwtID, accessTokenJwt.ExpiresAt) // starts a new goroutine which cleans this up after the expiry time from the map
	}
	return nil
}

// ? ----+-----+-----Forgot password-----+-----+-----

type forgotPasswordResult struct {
	channel     authChannel
	referenceID string
}

func (s *Service) forgotPassword(ctx context.Context, req forgotPasswordRequest) (forgotPasswordResult, error) {
	var channel authChannel
	var target string
	if req.Email != "" {
		channel = channelEmail
		target = req.Email
	} else if req.Phone != "" {
		channel = channelPhone
		target = req.Phone
	} else {
		return forgotPasswordResult{}, errMissingIdentifier
	}
	user, err := s.Repo.FindUserByChannel(ctx, channel, target)
	if err != nil {
		return forgotPasswordResult{}, err // db error
	}
	if user == nil {
		return forgotPasswordResult{}, errUserNotFound
	}

	otp, err := s.sendOTP(channel, purposeResetPass, target)
	refID := token.GenerateReferenceID()
	s.addOTPSession(refID, channel, purposeResetPass, user.ID, otp)
	return forgotPasswordResult{
		channel:     channel,
		referenceID: refID,
	}, nil
}
