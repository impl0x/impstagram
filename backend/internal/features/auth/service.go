package auth

import (
	"backend/internal/pkg/dob"
	"backend/internal/pkg/email"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/password"
	"backend/internal/pkg/token" // used to generate cryptographically random IDs and OTPs
	"backend/internal/pkg/ttlcache"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// some naming clarifications if it is confusing
// - Reset always means reset password

type Service struct {
	repo  repository
	email email.Sender
	otp   token.OtpGenerator

	mu            sync.RWMutex                          // mutex for the OTPSessions
	otpSessions   *ttlcache.Cache[string, otpSession]   // string is the reference id
	resetSessions *ttlcache.Cache[string, resetSession] // string is reference id
}

func NewService(repo repository, emailClient email.Sender) *Service {
	return &Service{
		email:         emailClient,
		repo:          repo,
		otpSessions:   ttlcache.New[string, otpSession](ruleTTLCacheCleanIntervalOTP),
		resetSessions: ttlcache.New[string, resetSession](ruleTTLCacheCleanIntervalReset),
	}
}

// ? ----+-----+-----Sessions-----+-----+-----

// Used to store the pending otp sessions
type otpSession struct {
	userID  uuid.UUID
	channel authChannel // e.g., channelEmail
	purpose authPurpose
	otp     string
}

// Used to store the pending reset password sessions
type resetSession struct {
	userID uuid.UUID
}

// ? ----+-----+-----Helper Wrapper Functions for token generation-----+-----+-----

// generates a new random string with the prefix [rulePrefixRefreshToken] and size [ruleSizeRefreshToken]
func generateRefreshToken() string {
	return token.GenerateToken(rulePrefixRefreshToken, ruleSizeRefreshToken)
}

// generates a new random string with the prefix [rulePrefixOTPSession] and size [ruleSizeSessionID]
func generateOTPSessionID() string {
	return token.GenerateToken(rulePrefixOTPSession, ruleSizeSessionID)
}

// generates a new random string with the prefix [rulePrefixResetSession] and size [ruleSizeSessionID]
func generateResetSessionID() string {
	return token.GenerateToken(rulePrefixResetSession, ruleSizeSessionID)
}

// ? ----+-----+-----Register-----+-----+-----

// Stores the result to the register call
type registerResult struct {
	referenceID string      // Used to link the upcoming OTP request
	channel     authChannel // the identifier/channel used to register, eg: email/phone
}

// Registers a new user
func (s *Service) register(ctx context.Context, req registerRequest) (registerResult, error) {
	// validate the date of birth string (this technically should be in the handler layer but then I would have to pass dob object separately in this function which would look bad so i just did it here)
	userDob, err := dob.NewDobFromString(req.Dob)
	switch err {
	case nil:
	case dob.ErrInvalidDobString:
		return registerResult{}, errInvalidDobString
	case dob.ErrImpossibleDob:
		return registerResult{}, errImpossibleDobString
	default:
		return registerResult{}, err
	}

	// validate the user's age against our business rules
	userAge := userDob.Age()
	if userAge < ruleMinAge {
		return registerResult{}, errNotOldEnough
	} else if userAge > ruleMaxAge {
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
	user, err = s.repo.findUserByChannel(ctx, channel, target)

	// if db returned a user
	if err == nil && user != nil {
		// i acknowledge that user has chances of being banned/unverified, but this is intended. We want user to login and then hit those errors if they exist.
		return registerResult{}, errAlreadyExistingUser
	} else if err != errRepoNoResults { // if the error received was not a no results error then must be a database error
		return registerResult{}, err
	}

	// checks if the username already exists because usernames are unique
	sameUsernameUser, err := s.repo.findUserByChannel(ctx, channelUsername, req.Username) // assuming req.Username is validated in validator
	if sameUsernameUser != nil || err == nil {                                            // todo: handle db level errors
		return registerResult{}, errUsernameAlreadyExists
	}

	// Now we hash the user's password and create a new user in the database
	user = newUser(req, password.Hash(req.Password)) // returns a unverified user by default
	user.ID, err = s.repo.createUser(ctx, user)      // we create a user before sending otp
	if err != nil {
		return registerResult{}, err
	}

	// Verify the identifier. By sending an otp
	otp, err := s.sendOTP(channel, purposeRegistration, target)
	if err != nil {
		return registerResult{}, err
	}
	// generate a new reference id for a otp session
	refID := generateOTPSessionID()
	// add a new otp session to our timed cache
	s.otpSessions.Add(
		refID,
		otpSession{
			userID:  user.ID,
			channel: channel,
			purpose: purposeRegistration,
			otp:     otp,
		},
		time.Now().Add(ruleExpiryTimeOTP),
	)

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

// Login a user in, and optionally if the user has 2fa enabled it asks for a code on the verify endpoint.
//
// rmd requestMetadata is required for userSession storage on successful login
func (s *Service) login(ctx context.Context, req loginRequest, rmd requestMetadata) (loginResult, error) {
	// Figure out what identifier the user sent, i.e. email/phone/username to log in
	var user *user
	var err error
	switch {
	case req.Username != "":
		user, err = s.repo.findUserByChannel(ctx, channelUsername, req.Username)
	case req.Email != "":
		user, err = s.repo.findUserByChannel(ctx, channelEmail, req.Email)
	case req.Phone != "":
		user, err = s.repo.findUserByChannel(ctx, channelPhone, req.Phone)
	default:
		return loginResult{}, errMissingIdentifierLogin
	}

	if err != nil {
		if err == errRepoNoResults {
			return loginResult{}, errCredentialsInvalid
		}
		return loginResult{}, err
	}

	// if user is found in database we compare the passwords and the password hash in the database to see if the user has the correct password
	ok, err := password.Compare(req.Password, user.PasswordHash)
	if err != nil {
		return loginResult{}, err
	}
	if !ok {
		return loginResult{}, errCredentialsInvalid
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
		refID := generateOTPSessionID()
		primaryTwoFAChannel := user.TwoFAs[0] // by default the first element is the primary 2FA identifier
		// if it is TOTP
		if primaryTwoFAChannel == channelTOTP { // if its a time based otp we don't bother generating or sending it anywhere

			s.otpSessions.Add( // we don't save an otp but instead save the secret key from the database to reduce a db call on the verify endpoint
				refID,
				otpSession{
					userID:  user.ID,
					channel: channelTOTP,
					purpose: purpose2FA,
					otp:     user.TotpSecretKey,
				},
				time.Now().Add(ruleExpiryTimeOTP),
			)

			return loginResult{
				requires2FA: true,
				channel:     channelTOTP,
				referenceID: refID,
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
		// Adding new otp session to our cache
		s.otpSessions.Add(
			refID,
			otpSession{
				userID:  user.ID,
				channel: primaryTwoFAChannel,
				purpose: purpose2FA,
				otp:     otp,
			},
			time.Now().Add(ruleExpiryTimeOTP),
		)
		return loginResult{
			requires2FA: true,
			channel:     primaryTwoFAChannel,
			referenceID: refID,
		}, nil
	}
	// else

	// generate tokens
	jwtID := uuid.New()
	accessToken, err := jwt.GenerateToken(jwt.NewAccessTokenPayload(user.ID, jwtID))
	if err != nil {
		panic(err)
	}
	refreshToken := generateRefreshToken()

	// Add a new user session to the database
	err = s.repo.createSession(
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

// ? ----+-----+-----Forgot password-----+-----+-----

type forgotPasswordResult struct {
	channel     authChannel
	referenceID string
}

func (s *Service) forgotPassword(ctx context.Context, req forgotPasswordRequest) (forgotPasswordResult, error) {
	// Figure out what identifier/channel the user sent us and store those into variables
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
	// Find the user in the database
	user, err := s.repo.findUserByChannel(ctx, channel, target)
	if err != nil {
		if err==errRepoNoResults{
			return forgotPasswordResult{}, errUserNotFound
		}
		return forgotPasswordResult{}, err // db error
	}

	// Send an otp to the channel and target our user sent us
	otp, err := s.sendOTP(channel, purposeResetPass, target)
	// generate a otp session id and add it to our ttlcache
	refID := generateOTPSessionID()
	s.otpSessions.Add(
		refID,
		otpSession{
			userID:  user.ID,
			channel: channel,
			purpose: purposeResetPass,
			otp:     otp,
		},
		time.Now().Add(ruleExpiryTimeOTP),
	)
	// send the session id as reference to the user
	return forgotPasswordResult{
		channel:     channel,
		referenceID: refID,
	}, nil
}

// ? ----+-----+-----Reset password-----+-----+-----

func (s *Service) resetPassword(ctx context.Context, req resetPasswordRequest) error {
	// Retrieve the reset session from the cache using the reference id from the request data
	session, expiresAt, ok := s.resetSessions.Get(req.ReferenceID)
	if !ok {
		return errResetSessionNotFound
	}
	// if the reset request expired we return error
	if expiresAt.Before(time.Now()) {
		return errResetSessionExpired
	}

	// else we proceed and update the user's password, we of course hash it.
	err := s.repo.updateUserPassword(ctx, session.userID, password.Hash(req.NewPassword))
	if err != nil {
		if err==errRepoNoResults{
			return errUserNotFound
		}
		return err 
	}
	// delete the session to make sure this reference id cannot be reused
	s.resetSessions.Delete(req.ReferenceID)
	return nil
}

// ? [HELPER] ----+-----+-----Send otp-----+-----+-----

// SendOTP sends a one-time password challenge to a user destination.
// The behavior of this function changes based on the following configurations:
//   - channel: The transport medium used to deliver the token (Email or SMS)
//   - purpose: The system context (2FA, Reset password or Registration) used to select templates
//   - target: The absolute address string (e.g., an email address or E.164 phone number)
func (s *Service) sendOTP(channel authChannel, purpose authPurpose, target string) (string, error) {
	otp, err := s.otp.Generate()
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
		err = s.email.Send(emailSendRequest)
	case channelPhone:
		// update: we can never send otp to phone numbers as its not possible as of now to afford sms service or telegram's gateway service.
		// will not change the code but anyone signing up on the backend with a phone will simply not receive an otp and will not be able to continue.
		// not returning an error or anything, its intentional, due to future compatibility. the frontend should not have phone supported.
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
	accessToken    string
	refreshToken   string
	isResetRequest bool   // if this is true the above two values are zero'd out
	referenceID    string // if this verify request was for a reset password we return a referenceID instead
}

// Verifies the two factor / verification / reset password OTP and generates a token / reset pass id for the user.
func (s *Service) verifyOTP(ctx context.Context, req verifyOTPRequest, rmd requestMetadata) (verifyResult, error) { // token, error
	// retrieve the session from the reference id in the request
	session, expiresAt, ok := s.otpSessions.Get(req.ReferenceID)
	if !ok { // if not found it means either the frontend is trying to reuse the same reference id after expiration or verification
		return verifyResult{}, errRefIDNotFound
	}
	if expiresAt.Before(time.Now()) { // if the otp has expired we return a error, the ttl cache automatically removes any values which are expired on a Get call if Cache.Config.LazyDelete is set to true, which is default.
		return verifyResult{}, errOTPExpired
	}
	var err error
	if session.channel == channelTOTP { // if its a authenticator time based 2 factor code
		session.otp, err = token.GenerateTOTP(session.otp) // in this case secretOTP is actually the secretKey for the TOTP which is used to compute the otp.
		if err != nil {
			return verifyResult{}, err
		}
	}
	if session.otp != req.OTP { // we check the otp if it does not match we return early with a incorrect otp error
		return verifyResult{}, errIncorrectOTP
	}
	// if the otp matches then we remove the reference id from our map immediately
	s.otpSessions.Delete(req.ReferenceID)
	// find the user
	user, err := s.repo.findUserByID(ctx, session.userID)
	if err != nil {
		if err == errRepoNoResults {
			return verifyResult{}, errUserNotFound // if the user mysteriously got deleted after just trying to log in or register...
		}
		return verifyResult{}, err
	}

	// Switch on the purpose to do purpose related tasks
	switch session.purpose {
	case purposeRegistration: // if registration we need to set user status to verified in the database
		err = s.repo.updateUserStatus(ctx, user.ID, statusVerified)
		if err != nil { // no need to handle for errRepoNoResults because we did that above
			return verifyResult{}, err
		}
	case purposeResetPass:
		refID := generateResetSessionID()
		s.resetSessions.Add(
			refID,
			resetSession{
				userID: user.ID,
			},
			time.Now().Add(ruleExpiryTimeResetPassword),
		)
		return verifyResult{
			isResetRequest: true,
			referenceID:    refID,
		}, nil
	}

	// generate new tokens and return them to the user for future usage
	jwtID := uuid.New()
	accessToken, err := jwt.GenerateToken(jwt.NewAccessTokenPayload(user.ID, jwtID))
	if err != nil {
		panic(err)
	}
	refreshToken := generateRefreshToken()

	// Add a new user session to the database
	err = s.repo.createSession(
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

func (s *Service) refresh(ctx context.Context, req refreshRequest) (refreshResult, error) {
	// Do a db lookup with the refresh token's hash
	userSesh, err := s.repo.findSessionByToken(ctx, token.GenerateMD5Hash(req.RefreshToken))
	if err != nil {
		return refreshResult{}, err // db error
	}
	if userSesh == nil {
		return refreshResult{}, errInvalidRefreshToken
	}

	// Check if the session has expired
	if userSesh.ExpiresAt.Before(time.Now()) {
		// Deleting the user session if it is expired, the user will have to create a new session again by logging in.
		err = s.repo.deleteSessionByID(ctx, userSesh.ID)
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
	refreshToken := generateRefreshToken()

	// update the session with the new refresh token and also update the expires at field to the max capacity again.
	err = s.repo.updateSessionToken(
		ctx,
		userSesh.ID,
		token.GenerateMD5Hash(refreshToken), // we store a hash of the token
		time.Now().AddDate(0, 0, ruleExpiryTimeRefreshToken),
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
	// Remove user session from database
	err := s.repo.deleteSessionByJwtID(ctx, accessTokenJwt.JwtID)
	if err != nil {
		return err // db error
	}
	// Add the jwt token id to the block list so this gets rejected by the auth middleware
	jwtTokenBlockList.Add(accessTokenJwt.JwtID, struct{}{}, accessTokenJwt.ExpiresAt)
	return nil
}
