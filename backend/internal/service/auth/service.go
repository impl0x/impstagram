package auth

import (
	"backend/internal/entity"     // entity data objects
	"backend/internal/repository" // repository
	"backend/internal/service"
	"backend/internal/service/auth/config" // config values
	"backend/pkg/cryptoutil"               // generate cryptographically random IDs and OTPs and hash passwords
	"backend/pkg/email"                    // send emails
	"backend/pkg/email/render"             // render emails
	"backend/pkg/jwt"                      // generate and verify jwt tokens
	"context"
	"crypto/subtle"
	"slices"
	"time"
	"uuid"

	"github.com/impl0x/go-utils/cache"  // ttlcache
	"github.com/impl0x/go-utils/logger" // logging
)

// ! INFO:
// main file for the core business logic for auth

// authPurpose defines WHY the OTP or action is happening.
type authPurpose string

const (
	purposeNil              authPurpose = ""
	purposeRegistration     authPurpose = "registration"
	purpose2FA              authPurpose = "2fa"
	purposeForgotPassword   authPurpose = "forgot_password"
	purposeChangeIdentifier authPurpose = "change_identifier"
)

type Service struct {
	cache    serviceCaches
	repo     Repositories
	otp      otpManager
	totp     cryptoutil.TOTPManager
	password cryptoutil.PasswordHasher
	email    _email
	token    token
	log      logger.Logger
	config   config.Config
}

type token struct {
	jwt                jwt.Manager       // access token
	refreshTokenHasher cryptoutil.Hasher // refresh token hasher function
}

type _email struct { // prefixed with underscore because "email" already exists as a package import
	sender   email.Sender
	renderer render.Renderer
}

type serviceCaches struct {
	jwtBlocklist   *cache.TTLCache[uuid.UUID, struct{}]
	forgotPassword *cache.TTLLeaseCache[string, forgotPasswordSession]
	totpSetup      *cache.TTLLeaseCache[string, *totpSetupSession]
} // some fields use pointer while others don't due to nature of modification to the struct, read only structs are passed by value.

// repositories needed
type Repositories struct {
	User    repository.UserRepository
	Session repository.UserSessionRepository
	Profile profileRepository
}

// segregated interface because we only need one of the methods available in [repository.profileRepository]
type profileRepository interface {
	CheckUsername(ctx context.Context, username string) (bool, error)
}

type Dependencies struct {
	Log                logger.Logger             // logging
	Config             config.Config             // configuration and business rules
	Repos              Repositories              // store and retrieve data
	Otp                cryptoutil.OTPManager     // generate otps
	Totp               cryptoutil.TOTPManager    // setup and generate totps
	Password           cryptoutil.PasswordHasher // hash and verify passwords
	EmailSender        email.Sender              // send emails
	EmailRenderer      *render.Renderer          // render email html templates with data
	Jwt                jwt.Manager               // manage access token
	RefreshTokenHasher cryptoutil.Hasher         // hash refresh token and save in user session repository
}

// Instantiates a new service instance with default rules and caches, and repositories provided in the parameters.
func NewService(d Dependencies) *Service {
	d.Log.Debug("Auth Service: created new instance")
	s := Service{
		log:    d.Log,
		config: d.Config,
		cache: serviceCaches{
			jwtBlocklist:   cache.NewTTLCache[uuid.UUID, struct{}](d.Config.TTLCacheCleanInterval.JWTBlockList),
			forgotPassword: cache.NewTTLLeaseCache[string, forgotPasswordSession](d.Config.TTLCacheCleanInterval.ForgotPassword),
			totpSetup:      cache.NewTTLLeaseCache[string, *totpSetupSession](d.Config.TTLCacheCleanInterval.TOTPSetup),
		},
		repo:     d.Repos,
		totp:     d.Totp,
		password: d.Password,
		token: token{
			jwt:                d.Jwt,
			refreshTokenHasher: d.RefreshTokenHasher,
		},
	}
	s.otp = otpManager{
		otp:                d.Otp,
		sessionIDGenerator: s.generateOTPSessionID,
		channels:           otpManagerChannels{&s.email},
		sessions:           cache.NewTTLLeaseCache[string, *otpSession](s.config.TTLCacheCleanInterval.OTP),
		expiryTime:         s.config.ExpiryTime.OTP,
		expiryTimeString:   timeDurationToString(s.config.ExpiryTime.OTP),
		maxAttempts:        s.config.MaxAttempts.Verify.OTP,
		maxResends:         s.config.MaxAttempts.Resend.OTP,
		log:                s.log,
	}
	return &s
}

// ? ----+-----+-----Cache sessions structs-----+-----+-----

// type totpSession struct {
// 	userID    uuid.UUID
// 	secretKey string
// 	attempts  uint8
// 	purpose   authPurpose
// }

// Used to store the pending reset password sessions
type forgotPasswordSession struct {
	userID uuid.UUID
}

// Used to store the pending totp sessions
type totpSetupSession struct {
	userID    uuid.UUID
	secretKey string
	attempts  uint8
}

// ? ----+-----+-----Wrapper functions for cryptoutil-----+-----+-----

// generates a new random string with the prefix [Service.config.PrefixRefreshToken] and size [Service.config.Size.RefreshToken]
func (s *Service) generateRefreshToken() string {
	return cryptoutil.GenerateToken(s.config.Prefix.RefreshToken, s.config.Size.RefreshToken)
}

// generates a new random string with the prefix [Service.config.Prefix.OTPSession] and size [Service.config.Size.SessionID]
func (s *Service) generateOTPSessionID() string {
	return cryptoutil.GenerateToken(s.config.Prefix.OTPSession, s.config.Size.SessionID)
}

// generates a new random string with the prefix [Service.config.Prefix.ForgotPasswordSession] and size [Service.config.Size.SessionID]
func (s *Service) generateForgotPasswordSessionID() string {
	return cryptoutil.GenerateToken(s.config.Prefix.ForgotPasswordSession, s.config.Size.SessionID)
}

// generates a new random string with the prefix [Service.config.Prefix.TOTPSession] and size [Service.config.Size.SessionID]
func (s *Service) generateTotpSessionID() string {
	return cryptoutil.GenerateToken(s.config.Prefix.TOTPSession, s.config.Size.SessionID)
}

// ? ----+-----+-----Helper functions-----+-----+-----

// generates new access and refresh tokens from the data provided and creates new user session in repository
//
// only errors that can be returned are from repository
func (s *Service) generateTokensAndCreateSession(ctx service.Context, user *entity.User, md entity.ClientMetadata) (string, string, error) {
	jwtID := uuid.New()
	// new access token
	accessToken := s.token.jwt.GenerateToken(
		newAccessToken( // generating a new access token struct and taking its payload version with json compatible tags
			user.ID,
			jwtID,
			s.config.ExpiryTime.AccessToken,
		).Claims(),
	)
	// new refresh token
	refreshToken := s.generateRefreshToken()
	// Add a new user session to the database
	session := &entity.UserSession{
		JwtID:     jwtID,
		UserID:    user.ID,
		TokenHash: s.token.refreshTokenHasher(refreshToken),
		ExpiresAt: time.Now().Add(s.config.ExpiryTime.RefreshToken),
	}

	// populating the optional fields
	if md.IPAddress != "" {
		session.IPAddress = &md.IPAddress
	}
	if md.OSName != "" {
		session.OSName = &md.OSName
	}
	if md.BrowserName != "" {
		session.BrowserName = &md.BrowserName
	}
	if md.DeviceType != "" {
		session.DeviceType = &md.DeviceType
	}

	// create user session in repository
	err := s.repo.Session.Create(ctx, session)

	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// ! ----x-----x-----NORMAL METHODS-----x-----x-----
// methods in service take a request data as a struct and return a result struct and error
// this pattern is common among all service methods

// ? ----+-----+-----Register-----+-----+-----

// Result struct
type RegisterResult struct {
	SessionID string    // Used to link the upcoming OTP request
	ExpiresAt time.Time // verification otp expiration
}

// Registers a new user
func (s *Service) Register(ctx service.Context, req RegisterRequest) (RegisterResult, error) {
	if !channelIsValidForRegister(req.Identity.Channel) { // basic validation to make sure we don't face unwanted panic later.
		panic("register: invalid channel found, channel - " + string(req.Identity.Channel))
	}

	// validate the user's age against our business rules
	userAge := req.Dob.Age()
	if userAge < s.config.Age.Min {
		return RegisterResult{}, errRegisterNotOldEnough
	} else if userAge > s.config.Age.Max {
		return RegisterResult{}, errRegisterTooOld
	}

	// Finding in database
	user, err := s.repo.User.GetByAuthIdentity(ctx, req.Identity)
	if err == nil && user != nil { // if no error and db returned a user
		// i acknowledge that user has chances of being banned/unverified, but this is intended. We want user to login and then hit those errors if they exist.
		return RegisterResult{}, errRegisterAlreadyExistingUser
	} else if err != repository.ErrNoResults { // if the error received was not a no results error then must be a database error
		return RegisterResult{}, err
	}

	// checks if the username already exists because usernames are unique/ this can also be done using database by directly inserting if the database has a unique tag for this field, but we still do it here.
	sameUsernameExists, err := s.repo.Profile.CheckUsername(ctx, req.Username) // assuming req.Username is validated in validator
	if err != nil && err != repository.ErrNoResults {
		return RegisterResult{}, err
	}
	if sameUsernameExists {
		return RegisterResult{}, errRegisterUsernameAlreadyExists
	}

	// creating user entity object (as of now by default status for user is 'unverified', so we do not need to set it ourselves)
	user = &entity.User{
		PasswordHash: s.password.Hash(req.Password), // hash password
		Dob:          req.Dob.String(),
	}
	// set user identifier value according to channel received
	switch req.Identity.Channel {
	case entity.ChannelEmail:
		user.Email = &req.Identity.Value
	case entity.ChannelPhone:
		user.Phone = &req.Identity.Value
	}
	// create user in both users and profiles table
	user.ID, err = s.repo.User.Create(ctx, user, req.Username)
	if err != nil {
		return RegisterResult{}, err
	}

	// send an welcome email if user registered via email
	// TODO: change this to support phone as-well
	if req.Identity.Channel == entity.ChannelEmail {
		// error is ignoreable in this case as we cannot really do anything if it happens.
		go s.email.sender.SendSingle( //nolint:errcheck
			ctx,
			s.email.renderer.Subject(render.SubWelcome),
			req.Identity.Value,
			s.email.renderer.Welcome(req.Username),
		)
	}

	sessionID, expiresAt, err := s.otp.send(ctx, req.Identity, user.ID, purposeRegistration)
	if err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	}, nil
}

// ? ----+-----+-----Check Username-----+-----+-----

type CheckUsernameResult struct {
	Exists bool
}

func (s *Service) CheckUsername(ctx service.Context, req CheckUsernameRequest) (CheckUsernameResult, error) {
	exists, err := s.repo.Profile.CheckUsername(ctx, req.Username)
	if err != nil {
		return CheckUsernameResult{}, err
	}
	return CheckUsernameResult{exists}, nil

}

// ? ----+-----+-----Login-----+-----+-----

// Stores the result to the login call
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	Requires2FA  bool               // if this is true then below fields are populated and must be returned with the error
	SessionID    string             // Used to link the upcoming OTP request
	ExpiresAt    time.Time          // otp expiration time
	Channel      entity.AuthChannel // the channel from which the otp we need to verify 2FA
}

// Login a user in, and optionally if the user has 2fa enabled it asks for a code on the verify endpoint.
func (s *Service) Login(ctx service.Context, req LoginRequest, md entity.ClientMetadata) (LoginResult, error) {
	if !channelIsValidForLogin(req.Identity.Channel) {
		panic("login: invalid channel found, channel - " + string(req.Identity.Channel))
	}
	user, err := s.repo.User.GetByAuthIdentity(ctx, req.Identity)
	if err != nil {
		if err == repository.ErrNoResults {
			return LoginResult{}, errLoginCredentialsInvalid
		}
		return LoginResult{}, err
	}

	// if user is found in database we compare the passwords and the password hash in the database to see if the user has the correct password
	ok, err := s.password.Compare(req.Password, user.PasswordHash)
	if err != nil {
		return LoginResult{}, err
	}
	if !ok {
		return LoginResult{}, errLoginCredentialsInvalid
	}
	// check if user is banned then we return immediately not allowing a login.
	switch user.Status {
	case entity.StatusBanned:
		return LoginResult{}, errLoginUserBanned
	case entity.StatusUnverified:
		return LoginResult{}, errLoginUserUnverified // frontend should hit the resend otp endpoint on this case
	}

	// If user has 2FA enabled ask for otp. then handle that
	if user.TwoFAs != nil {
		// the first element is the primary 2FA identifier
		primaryTwoFAChannel := user.TwoFAs[0]

		// session id and when the otp will expire variables that we need to send to the user
		var sessionID string
		var expiresAt time.Time

		// perform action according to what the channel is
		switch primaryTwoFAChannel {
		// if its TOTP, that is user has his primary 2FA as time based authentication code
		case entity.ChannelTOTP:
			if user.TotpSecretKey == nil {
				panic("login: User has 2FA in twoFAs slice but no secret key in TotpSecretKey")
			}
			// in this case we cannot "send" an otp, therefore we will save
			// a session directly in our otp cache, otp field will be empty and
			// identity will be made using the channel (TOTP) in this case and
			// value will be the secret key for the totp generation, which will
			// later be used at the verify endpoint.
			sessionID, expiresAt = s.otp.save(
				"",
				entity.AuthIdentity{
					Channel: primaryTwoFAChannel,
					Value:   *user.TotpSecretKey,
				},
				user.ID,
				purpose2FA,
			)

		// if its email we just send an otp
		case entity.ChannelEmail:
			if user.Email == nil {
				panic("login: user has primary 2fa as email but no email present in user entity")
			}
			sessionID, expiresAt, err = s.otp.send(
				ctx,
				entity.AuthIdentity{
					Channel: primaryTwoFAChannel,
					Value:   *user.Email,
				},
				user.ID,
				purpose2FA,
			)
			if err != nil {
				return LoginResult{}, err
			}
		// same as above
		case entity.ChannelPhone:
			if user.Phone == nil {
				panic("login: user has primary 2fa as phone but no phone present in user entity")
			}
			sessionID, expiresAt, err = s.otp.send(
				ctx,
				entity.AuthIdentity{
					Channel: primaryTwoFAChannel,
					Value:   *user.Phone,
				},
				user.ID,
				purpose2FA,
			)
			if err != nil {
				return LoginResult{}, err
			}
		}
		// return session id, expiry time and the channel for the otp.
		// with an error, because this is a error technically.
		return LoginResult{
			Requires2FA: true,
			SessionID:   sessionID,
			ExpiresAt:   expiresAt,
			Channel:     primaryTwoFAChannel,
		}, errLoginRequires2FA
	}
	// else if 2fa is not enabled we go on to generate the tokens

	// generate tokens
	accessToken, refreshToken, err := s.generateTokensAndCreateSession(ctx, user, md)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ? ----+-----+-----Resend OTP-----+-----+-----

// Resends an otp from a active session
func (s *Service) ResendOTP(ctx service.Context, req ResendOTPRequest) error {
	err := s.otp.resend(ctx, req.SessionID)
	if err != nil {
		return err
	}
	return nil
}

// ? ----+-----+-----Verify otp - multiple methods-----+-----+-----

// common struct to hold common fields
type verifyOTPResult struct {
	RemainingAttempts uint8 // only populated on specific otp incorrect error
}

type VerifyOTPForRegisterResult struct {
	AccessToken  string
	RefreshToken string
	verifyOTPResult
}

func (s *Service) VerifyOTPForRegister(ctx service.Context, req VerifyOTPRequest, md entity.ClientMetadata) (VerifyOTPForRegisterResult, error) {
	session, ra, err := s.otp.verify(req.SessionID, req.OTP, purposeRegistration)
	if err != nil {
		return VerifyOTPForRegisterResult{RemainingAttempts: ra}, err
	}
	// find the user
	user, err := s.repo.User.GetByID(ctx, session.userID)
	if err != nil {
		return VerifyOTPForRegisterResult{}, handleUserNotFound(err)
	}

	// change user status from unverified to verified
	err = s.repo.User.UpdateStatus(ctx, user.ID, entity.StatusVerified)
	if err != nil {
		return VerifyOTPForRegisterResult{}, err
	}

	// generate tokens and send them
	accessToken, refreshToken, err := s.generateTokensAndCreateSession(ctx, user, md)
	if err != nil {
		return VerifyOTPForRegisterResult{}, err
	}

	return VerifyOTPForRegisterResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

type VerifyOTPFor2FAResult struct {
	VerifyOTPForRegisterResult // embedded as both have the exact same fields
}

func (s *Service) VerifyOTPFor2FA(ctx service.Context, req VerifyOTPRequest, md entity.ClientMetadata) (VerifyOTPFor2FAResult, error) {
	session, ra, err := s.otp.verify(req.SessionID, req.OTP, purpose2FA)
	if err != nil {
		return VerifyOTPFor2FAResult{RemainingAttempts: ra}, err
	}
	// find the user
	user, err := s.repo.User.GetByID(ctx, session.userID)
	if err != nil {
		return VerifyOTPFor2FAResult{}, handleUserNotFound(err)
	}

	// generate tokens and send them
	accessToken, refreshToken, err := s.generateTokensAndCreateSession(ctx, user, md)
	if err != nil {
		return VerifyOTPFor2FAResult{}, err
	}

	return VerifyOTPFor2FAResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

type VerifyOTPForForgotPasswordResult struct {
	SessionID string
	ExpiresAt time.Time
	verifyOTPResult
}

func (s *Service) VerifyOTPForForgotPassword(ctx service.Context, req VerifyOTPRequest) (VerifyOTPForForgotPasswordResult, error) {
	session, ra, err := s.otp.verify(req.SessionID, req.OTP, purposeForgotPassword)
	if err != nil {
		return VerifyOTPForForgotPasswordResult{RemainingAttempts: ra}, err
	}
	// find the user
	user, err := s.repo.User.GetByID(ctx, session.userID)
	if err != nil {
		return VerifyOTPForForgotPasswordResult{}, handleUserNotFound(err)
	}

	// create new session for reset password and add to cache
	sessionID := s.generateForgotPasswordSessionID()
	expiresAt := time.Now().Add(s.config.ExpiryTime.ForgotPassword)
	s.cache.forgotPassword.Add(
		sessionID,
		forgotPasswordSession{
			userID: user.ID,
		},
		expiresAt,
	)
	// return the user the session id and expiry time
	return VerifyOTPForForgotPasswordResult{
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	}, nil
}

type VerifyOTPForChangeIdentifierResult struct {
	verifyOTPResult
}

func (s *Service) VerifyOTPForChangeIdentifier(ctx service.Context, req VerifyOTPRequest) (VerifyOTPForChangeIdentifierResult, error) {
	// verify the session otp
	session, ra, err := s.otp.verify(req.SessionID, req.OTP, purposeChangeIdentifier)
	if err != nil {
		return VerifyOTPForChangeIdentifierResult{RemainingAttempts: ra}, err
	}
	// update the identifier according to channel
	switch session.identity.Channel {
	case entity.ChannelEmail:
		err = s.repo.User.UpdateEmail(ctx, session.userID, session.identity.Value)
	case entity.ChannelPhone:
		err = s.repo.User.UpdatePhone(ctx, session.userID, session.identity.Value)
	default:
		panic("invalid identifier") // should not happen
	}
	if err != nil {
		return VerifyOTPForChangeIdentifierResult{}, err
	}
	return VerifyOTPForChangeIdentifierResult{}, nil

}

// ? ----+-----+-----Forgot password-----+-----+-----

type ForgotPasswordResult struct {
	SessionID string
	ExpiresAt time.Time
}

// Raises a forgot password session request which sends an verification otp to the channel provided and stores a in memory temporary session
func (s *Service) ForgotPassword(ctx service.Context, req ForgotPasswordRequest) (ForgotPasswordResult, error) {
	// Find the user in the database
	user, err := s.repo.User.GetByAuthIdentity(ctx, req.Identity)
	if err != nil {
		return ForgotPasswordResult{}, handleUserNotFound(err)
	}

	// Send an otp to the channel and target our user sent us
	sessionID, expiresAt, err := s.otp.send(ctx, req.Identity, user.ID, purposeForgotPassword)
	if err != nil {
		return ForgotPasswordResult{}, err
	}

	// send the session id as reference to the user
	return ForgotPasswordResult{
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	}, nil
}

// ? ----+-----+-----Reset password-----+-----+-----

// updates a user's password using a forgot password session
func (s *Service) ResetPassword(ctx service.Context, req ResetPasswordRequest) error {
	// Retrieve the reset session from the cache using the reference id from the request data
	item, ok := s.cache.forgotPassword.Acquire(req.SessionID)
	if !ok {
		return errResetPasswordSessionNotFound
	}
	defer item.Release()

	session, expiresAt := item.Value(), item.ExpiresAt()

	// if the reset request expired we return error
	if expiresAt.Before(time.Now()) {
		item.Delete()
		return errResetPasswordSessionExpired
	}

	// else we proceed and update the user's password, we of course hash it.
	err := s.repo.User.UpdatePasswordHash(ctx, session.userID, s.password.Hash(req.NewPassword))
	if err != nil {
		return handleUserNotFound(err)
	}
	// delete the session to make sure this session cannot be reused
	item.Delete()
	return nil
}

// ? ----+-----+-----Refresh-----+-----+-----

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

// refreshes the token pair with a new pair of tokens
func (s *Service) Refresh(ctx service.Context, req RefreshRequest) (RefreshResult, error) {
	// Do a db lookup with the refresh token's hash
	userSesh, err := s.repo.Session.GetByTokenHash(ctx, s.token.refreshTokenHasher(req.RefreshToken))
	if err != nil {
		if err == repository.ErrNoResults {
			return RefreshResult{}, errRefreshTokenInvalid
		}
		return RefreshResult{}, err
	}
	// Check if the session has expired
	if userSesh.ExpiresAt.Before(time.Now()) {
		// Deleting the user session if it is expired, the user will have to create a new session again by logging in.
		err = s.repo.Session.DeleteByID(ctx, userSesh.ID)
		if err != nil {
			return RefreshResult{}, err
		}
		return RefreshResult{}, errRefreshTokenExpired
	}

	// if everything is good we generate both new tokens, we do not have to regenerate and re update the jwt id as its unnecessary.
	// it is a fixed value which is linked with the user session in database
	accessToken := s.token.jwt.GenerateToken(
		newAccessToken(
			userSesh.UserID,
			userSesh.JwtID,
			s.config.ExpiryTime.AccessToken,
		).Claims(),
	)
	refreshToken := s.generateRefreshToken()

	// update the session with the new refresh token and also update the expires at field to the max capacity again.
	err = s.repo.Session.UpdateTokenHashAndExpiry(
		ctx,
		userSesh.ID,
		s.token.refreshTokenHasher(refreshToken), // we store a hash of the token
		time.Now().Add(s.config.ExpiryTime.RefreshToken),
	)
	if err != nil {
		if err == repository.ErrNoResults { // if the user logged out instantly somehow
			return RefreshResult{}, errRefreshTokenInvalid
		}
		return RefreshResult{}, err
	}

	// return both tokens
	return RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ! ----x-----x-----[AUTH TOKEN REQUIRED METHODS]-----x-----x-----

// ? ----+-----+-----Logout-----+-----+-----

// deletes the current user session
func (s *Service) Logout(ctx service.Context, token AccessToken) error {
	// Remove user session from database
	err := s.repo.Session.DeleteByJwtID(ctx, token.JwtID)
	if err != nil && err != repository.ErrNoResults { // ignore no results
		return err // db error
	}
	// Add the jwt token id to the block list so this gets rejected by the authorization
	s.cache.jwtBlocklist.Add(token.JwtID, struct{}{}, token.ExpiresAt)
	return nil
}

// ? ----+-----+-----Delete Account-----+-----+-----
func (s *Service) DeleteAccount(ctx service.Context, token AccessToken) error {
	err := s.repo.User.Delete(ctx, token.UserID) // assuming there is cascade such that profiles and user sessions also gets deleted.
	if err != nil {
		if err == repository.ErrNoResults {
			return errCommonUserNotFound
		}
		return err
	}
	return nil
}

// ? ----+-----+-----Add 2FA-----+-----+-----

// adds a 2FA method to the user account,
//
// note: totp has its separate function therefore its assumed that the channel is validated to be either email or phone only
func (s *Service) Add2FA(ctx service.Context, token AccessToken, req Add2FARequest) error {
	// finding the user in the database using the user id from token
	user, err := s.repo.User.GetByID(ctx, token.UserID)
	if err != nil {
		return handleUserNotFound(err)
	}
	// checking if the channel already has a 2fa
	channel := entity.AuthChannel(req.Channel) // assuming its validated to valid 2fa channels only, except totp it has its own endpoints
	if slices.Contains(user.TwoFAs, channel) {
		return errAdd2FAChannelExists
	}
	// switching on channel to check if the channel that the user requested is even present in our database
	switch channel {
	case entity.ChannelEmail:
		if user.Email == nil {
			return errAdd2FAChannelEmpty
		}
	case entity.ChannelPhone:
		if user.Phone == nil {
			return errAdd2FAChannelEmpty
		}
	}
	// adding the new channel to the 2fa list and updating in the database
	user.TwoFAs = append(user.TwoFAs, channel)
	err = s.repo.User.UpdateTwoFAs(ctx, user.ID, user.TwoFAs)
	if err != nil {
		return err
	}
	// returning empty result as there is nothing more we need to signify
	return nil
}

// Removes 2fa for a user
func (s *Service) Remove2FA(ctx service.Context, token AccessToken, req Remove2FARequest) error {
	// finding the user in the database using user id from the token
	user, err := s.repo.User.GetByID(ctx, token.UserID)
	if err != nil {
		return handleUserNotFound(err)
	}
	channel := entity.AuthChannel(req.Channel)
	// taking the index of the channel in the 2fa slice
	if user.TwoFAs == nil {
		return errRemove2FANotEnabled
	}
	// finding the index of the channel in the slice
	index := slices.Index(user.TwoFAs, channel)
	if index == -1 { // if not found
		return errRemove2FAChannelNotFound
	}
	// update the slice to not include the current channel
	user.TwoFAs = append(user.TwoFAs[:index], user.TwoFAs[index+1:]...)
	if len(user.TwoFAs) == 0 { // if this was the last 2fa then we disable 2fa altogether by setting the slice value to nil
		user.TwoFAs = nil
	}
	// update in the database
	err = s.repo.User.UpdateTwoFAs(ctx, user.ID, user.TwoFAs)
	if err != nil {
		if err == repository.ErrNoResults {
			return errCommonUserNotFound
		}
		return err
	}
	return nil
}

// ? ----+-----+-----Totp Setup-----+-----+-----

type TotpSetupResult struct {
	SessionID string
	TotpUri   string
	ExpiresAt time.Time
}

// starts a setup for totp
func (s *Service) TotpSetup(ctx service.Context, token AccessToken) (TotpSetupResult, error) {
	// Find user on the database
	user, err := s.repo.User.GetByID(ctx, token.UserID)
	if err != nil {
		return TotpSetupResult{}, handleUserNotFound(err)
	}
	// if user already has a totp secret key it means totp 2fa is enabled
	if user.TotpSecretKey != nil {
		return TotpSetupResult{}, errTotpSetupAlreadyEnabled // user needs to disable totp first to set it up again
	}
	// choose a identifier for the totp uri, preference being email
	var identifier string
	if user.Email != nil {
		identifier = *user.Email
	} else if user.Phone != nil {
		identifier = *user.Phone
	}
	// generating the secret key and uri
	secretKey, totpUri := s.totp.Setup(identifier)
	seshId := s.generateTotpSessionID()
	expiresAt := time.Now().Add(s.config.ExpiryTime.TOTPSetup)
	// adding to cache session
	s.cache.totpSetup.Add(
		seshId,
		&totpSetupSession{
			userID:    user.ID,
			secretKey: secretKey,
		},
		expiresAt,
	)
	return TotpSetupResult{
		SessionID: seshId,
		TotpUri:   totpUri,
		ExpiresAt: expiresAt,
	}, nil
}

// ? ----+-----+-----Totp Verify-----+-----+-----

type TotpVerifyResult struct {
	RemainingAttempts uint8
}

// verifies the otp from totp setup and sets secret in database
func (s *Service) TotpSetupVerify(ctx service.Context, token AccessToken, req TotpVerifyRequest) (TotpVerifyResult, error) {
	// Fetching session from cache
	item, ok := s.cache.totpSetup.Acquire(req.SessionID)
	if !ok {
		return TotpVerifyResult{}, errTotpSetupVerifySessionNotFound
	}
	defer item.Release()
	session, expiresAt := item.Value(), item.ExpiresAt()

	// checking if session and token user's match, otherwise it is a stolen token/reference id
	if session.userID != token.UserID {
		return TotpVerifyResult{}, errTotpSetupVerifySessionNotFound
	}
	// checking if session has expired
	if expiresAt.Before(time.Now()) {
		item.Delete()
		return TotpVerifyResult{}, errTotpSetupVerifySessionExpired
	}
	// calculating the otp
	otp, err := s.totp.Generate(session.secretKey)
	if err != nil {
		return TotpVerifyResult{}, err
	}
	session.attempts++ // increment on every attempt
	remainingAttempts := s.config.MaxAttempts.Verify.TOTPSetup - session.attempts
	// comparing against the otp sent
	if remainingAttempts <= 0 {
		item.Delete() // delete the session not allowing for any more verification attempts
		return TotpVerifyResult{}, errCommonAttemptsExhausted
	}
	// compare both otps
	if subtle.ConstantTimeCompare([]byte(req.OTP), []byte(otp)) == 1 {
		return TotpVerifyResult{RemainingAttempts: remainingAttempts}, errTotpSetupVerifyTOTPIncorrect
	}
	// delete the session if otp matches and verification is complete
	item.Delete()
	// enable totp in database
	err = s.repo.User.UpdateAddTotp(ctx, session.userID, session.secretKey)
	if err != nil {
		return TotpVerifyResult{}, err
	}
	return TotpVerifyResult{}, nil
}

// changes the users password
func (s *Service) ChangePassword(ctx service.Context, token AccessToken, req ChangePasswordRequest) error {
	user, err := s.repo.User.GetByID(ctx, token.UserID)
	if err != nil {
		return err
	}
	isPwSame, err := s.password.Compare(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		return err
	}
	if !isPwSame {
		return errChangePasswordWrongCurrentPassword
	}
	err = s.repo.User.UpdatePasswordHash(ctx, token.UserID, s.password.Hash(req.NewPassword))
	if err != nil {
		return err
	}
	return nil
}

type ChangeIdentifierResult struct {
	SessionID string
	ExpiresAt time.Time
}

// changes the user's identifier, either email or phone
func (s *Service) ChangeIdentifier(ctx service.Context, token AccessToken, req ChangeIdentifierRequest) (ChangeIdentifierResult, error) {
	sessionID, expiresAt, err := s.otp.send(ctx, req.Identity, token.UserID, purposeChangeIdentifier)
	if err != nil {
		return ChangeIdentifierResult{}, err
	}
	return ChangeIdentifierResult{sessionID, expiresAt}, nil
}
