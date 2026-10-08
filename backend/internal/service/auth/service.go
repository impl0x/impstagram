package auth

import (
	"backend/internal/config"
	"backend/internal/entity"     // entity data objects
	"backend/internal/repository" // repository
	"backend/pkg/cryptoutil"      // generate cryptographically random IDs and OTPs
	"backend/pkg/email"           // send emails
	"backend/pkg/jwt"             // generate and verify jwt tokens
	"backend/pkg/password"        // hash and compare passwords
	"context"
	"fmt"
	"slices"
	"time"
	"uuid"

	"github.com/impl0x/go-utils/cache" // used for ttlcache
)

// ! INFO:
// main file for the core business logic for auth

// authPurpose defines WHY the OTP or action is happening.
type authPurpose string

const (
	purposeRegistration authPurpose = "registration"
	purpose2FA          authPurpose = "2fa"
	purposeResetPass    authPurpose = "reset_password"
)

type Service struct {
	config config.AuthConfig
	cache  serviceCaches
	repo   repositories
	otp    cryptoutil.OtpGenerator
	jwt    jwt.JWTManager
	// TODO: add a proper logger
	// log   *log.Logger

	email email.Sender
}

type serviceCaches struct {
	otp           *cache.TTLCache[string, *otpSession]
	resetPassword *cache.TTLCache[string, resetPasswordSession]
	totp          *cache.TTLCache[string, *totpSession]
	jwtBlocklist  *cache.TTLCache[uuid.UUID, struct{}]
} // some fields use pointer while others don't due to nature of modification to the struct, read only structs are passed by value.

type repositories struct {
	user    repository.UserRepository
	session repository.UserSessionRepository
	profile repository.ProfileRepository
}

// wrapper to pass repositories into [NewService] easier
func NewRepositories(user repository.UserRepository, session repository.UserSessionRepository, profile repository.ProfileRepository) repositories {
	return repositories{user, session, profile}
}

// Instantiates a new service instance with default rules and caches, and repositories provided in the parameters.
func NewService(
	serviceName string,
	cfg *config.AuthConfig,
	repos repositories,
	emailClient email.Sender,
) *Service {
	if repos.user == nil || repos.session == nil || repos.profile == nil || emailClient == nil {
		panic("Service: nil parameter found in instantiation function")
	}
	config := *cfg // dereferencing pointer to store as whole in struct
	return &Service{
		config: config,
		cache: serviceCaches{
			otp:           cache.NewTTLCache[string, *otpSession](config.TTLCacheCleanIntervalOTP),
			resetPassword: cache.NewTTLCache[string, resetPasswordSession](config.TTLCacheCleanIntervalReset),
			totp:          cache.NewTTLCache[string, *totpSession](config.TTLCacheCleanIntervalTOTP),
			jwtBlocklist:  cache.NewTTLCache[uuid.UUID, struct{}](config.TTLCacheCleanIntervalJWTBlockList),
		},
		repo:  repos,
		otp:   cryptoutil.NewOtpGenerator(serviceName, config.LenOTP, config.LenTOTP, cfg.SizeTOTPKey),
		jwt:   jwt.NewJWTManager(config.JwtSecret),
		email: emailClient,
	}
}

// ? ----+-----+-----Cache sessions structs-----+-----+-----

// Used to store the pending otp sessions
type otpSession struct {
	userID   uuid.UUID
	channel  entity.AuthChannel // e.g., ChannelEmail
	purpose  authPurpose        // e.g., PurposeLogin
	otp      string
	attempts int
}

// Used to store the pending reset password sessions
type resetPasswordSession struct {
	userID uuid.UUID
}

// Used to store the pending totp sessions
type totpSession struct {
	userID    uuid.UUID
	secretKey string
	attempts  int
}

// ? ----+-----+-----Wrapper functions for cryptoutil-----+-----+-----

// generates a new random string with the prefix [Service.config.PrefixRefreshToken] and size [Service.config.SizeRefreshToken]
func (s *Service) generateRefreshToken() string {
	return cryptoutil.GenerateToken(s.config.PrefixRefreshToken, s.config.SizeRefreshToken)
}

// generates a new random string with the prefix [Service.config.PrefixOTPSession] and size [Service.config.SizeSessionID]
func (s *Service) generateOTPSessionID() string {
	return cryptoutil.GenerateToken(s.config.PrefixOTPSession, s.config.SizeSessionID)
}

// generates a new random string with the prefix [Service.config.PrefixResetSession] and size [Service.config.SizeSessionID]
func (s *Service) generateResetSessionID() string {
	return cryptoutil.GenerateToken(s.config.PrefixResetSession, s.config.SizeSessionID)
}

// generates a new random string with the prefix [Service.config.PrefixTOTPSession] and size [Service.config.SizeSessionID]
func (s *Service) generateTotpSessionID() string {
	return cryptoutil.GenerateToken(s.config.PrefixTOTPSession, s.config.SizeSessionID)
}

// ? ----+-----+-----Helper functions-----+-----+-----

// SendOTP sends a one-time password challenge to a user destination.
// The behavior of this function changes based on the following configurations:
//   - channel: The transport medium used to deliver the OTP (Email or SMS)
//   - purpose: The system context (2FA, Reset password or Registration) used to select templates
//   - target: The absolute address string (e.g., an email address or E.164 phone number)
//
// it returns the otp that was sent and an optional error occurs one
func (s *Service) sendOTP(channel entity.AuthChannel, purpose authPurpose, target string) (string, error) {
	otp := s.otp.Generate()
	var err error
	// Send otp based on the identifier
	switch channel {
	case entity.ChannelEmail:
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
	case entity.ChannelPhone:
		// IMPORTANT
		// we can never send otp to phone numbers as its not possible as of now to afford sms service or telegram's gateway service.
		// will not change the code but anyone signing up on the backend with a phone will simply not receive an otp and will not be able to continue.
		// not returning an error or anything, its intentional, due to future compatibility. the frontend should not have phone supported.
	default:
		panic("send otp: invalid channel passed: " + string(channel))
	}
	if err != nil { // send otp error
		return "", err
	}
	return otp, nil
}

// helper function to create a new user session entity using client metadata and other required parameters.
//
// by default hashes the refresh token using [cryptoutil.GenerateMD5Hash]
func newUserSession(userID, jwtID uuid.UUID, refreshToken string, md entity.ClientMetadata) *entity.UserSession {
	session := &entity.UserSession{
		JwtID:     jwtID,
		UserID:    userID,
		TokenHash: cryptoutil.GenerateMD5Hash(refreshToken),
	}
	// populating struct fields according to what's populated in client metadata struct
	if md.IPAddress != "" {
		session.IPAddress = &md.IPAddress
	}
	if md.UserAgent.OSName != "" {
		session.OSName = &md.UserAgent.OSName
	}
	if md.UserAgent.BrowserName != "" {
		session.BrowserName = &md.UserAgent.BrowserName
	}
	if md.UserAgent.DeviceType != "" {
		session.DeviceType = &md.UserAgent.DeviceType
	}
	return session
}

// ! ----x-----x-----NORMAL METHODS-----x-----x-----
// methods in service take a request data as a struct and return a result struct and error
// this pattern is common among all service methods

// ? ----+-----+-----Register-----+-----+-----

// Result struct
type RegisterResult struct {
	ReferenceID string    // Used to link the upcoming OTP request
	ExpiresAt   time.Time // verification otp expiration
}

// Registers a new user
func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResult, error) {
	if !channelIsValidForRegister(req.Channel) { // basic validation to make sure we don't face unwanted panic later.
		panic("register: invalid channel found")
	}

	// validate the user's age against our business rules
	userAge := req.Dob.Age()
	if userAge < s.config.AgeMin {
		return RegisterResult{}, errRegisterNotOldEnough
	} else if userAge > s.config.AgeMax {
		return RegisterResult{}, errRegisterTooOld
	}

	// Finding in database
	user, err := s.repo.user.GetByAuthChannel(ctx, req.Channel, req.Value)

	if err == nil && user != nil { // if no error and db returned a user
		// i acknowledge that user has chances of being banned/unverified, but this is intended. We want user to login and then hit those errors if they exist.
		return RegisterResult{}, errRegisterAlreadyExistingUser
	} else if err != repository.ErrNoResults { // if the error received was not a no results error then must be a database error
		return RegisterResult{}, err
	}

	// checks if the username already exists because usernames are unique/ this can also be done using database by directly inserting if the database has a unique tag for this field, but we still do it here.
	sameUsernameUser, err := s.repo.user.GetByAuthChannel(ctx, entity.ChannelUsername, req.Username) // assuming req.Username is validated in validator
	if sameUsernameUser != nil || err == nil {
		return RegisterResult{}, errRegisterUsernameAlreadyExists
	}
	// creating user entity object (as of now by default status for user is 'unverified', so we do not need to set it ourselves)
	user = &entity.User{
		PasswordHash: password.Hash(req.Password), // hash password
		Dob:          req.Dob.String(),
	}
	// set user identifier value according to channel received
	switch req.Channel {
	case entity.ChannelEmail:
		user.Email = &req.Value
	case entity.ChannelPhone:
		user.Phone = &req.Value
	}
	// create user in both users and profiles table
	user.ID, err = s.repo.user.Create(ctx, user, req.Username)
	if err != nil {
		return RegisterResult{}, err
	}

	// Verify the identifier. By sending an otp
	otp, err := s.sendOTP(req.Channel, purposeRegistration, req.Value)
	if err != nil {
		return RegisterResult{}, err
	}
	// generate a new reference id for a otp session
	refID := s.generateOTPSessionID()
	expiresAt := time.Now().Add(s.config.ExpiryTimeOTP)
	// add a new otp session to our timed cache
	s.cache.otp.Add(
		refID,
		&otpSession{
			userID:  user.ID,
			channel: req.Channel,
			purpose: purposeRegistration,
			otp:     otp,
		},
		expiresAt,
	)

	return RegisterResult{
		ReferenceID: refID,
		ExpiresAt:   expiresAt,
	}, nil
}

// ? ----+-----+-----Check Username-----+-----+-----

type CheckUsernameResult struct {
	Exists bool
}

func (s *Service) CheckUsername(ctx context.Context, req CheckUsernameRequest) (CheckUsernameResult, error) {
	exists, err := s.repo.profile.CheckUsername(ctx, req.Username)
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
	Requires2FA  bool   // if this is false then below all fields are zeroed out, else the above tokens is zero valued
	ReferenceID  string // Used to link the upcoming OTP request
	ExpiresAt    time.Time
}

// Login a user in, and optionally if the user has 2fa enabled it asks for a code on the verify endpoint.
//
// rmd requestMetadata is required for userSession storage on successful login
func (s *Service) Login(ctx context.Context, req LoginRequest, md entity.ClientMetadata) (LoginResult, error) {
	if !channelIsValidForLogin(req.Channel) {
		panic("login: invalid channel found")
	}
	user, err := s.repo.user.GetByAuthChannel(ctx, req.Channel, req.Value)
	if err != nil {
		if err == repository.ErrNoResults {
			return LoginResult{}, errLoginCredentialsInvalid
		}
		return LoginResult{}, err
	}

	// if user is found in database we compare the passwords and the password hash in the database to see if the user has the correct password
	ok, err := password.Compare(req.Password, user.PasswordHash)
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

	// If user has 2FA enabled ask for otp.
	if user.TwoFAs != nil {
		refID := s.generateOTPSessionID()
		expiresAt := time.Now().Add(s.config.ExpiryTimeOTP)
		primaryTwoFAChannel := user.TwoFAs[0] // by default the first element is the primary 2FA identifier
		// if it is TOTP
		if primaryTwoFAChannel == entity.ChannelTOTP { // if its a time based otp we don't bother generating or sending it anywhere
			if user.TotpSecretKey == nil {
				panic("login: User has 2FA in twoFAs slice but no secret key in TotpSecretKey")
			}
			s.cache.otp.Add( // we don't save an otp but instead save the secret key from the database to reduce a db call on the verify endpoint
				refID,
				&otpSession{
					userID:  user.ID,
					channel: entity.ChannelTOTP,
					purpose: purpose2FA,
					otp:     *user.TotpSecretKey,
				},
				expiresAt,
			)

			return LoginResult{
				Requires2FA: true,
				ReferenceID: refID,
				ExpiresAt:   expiresAt,
			}, nil
		}
		// else if its email or phone
		var target string // either the email or the phone literal
		switch primaryTwoFAChannel {
		case entity.ChannelEmail:
			if user.Email == nil {
				panic("login: user has primary 2fa as email but no email present in user entity")
			}
			target = *user.Email
		case entity.ChannelPhone:
			if user.Phone == nil {
				panic("login: user has primary 2fa as phone but no email present in user entity")
			}
			target = *user.Phone
		default:
			panic("login: invalid channel found in 2fa slice of user")
		}
		otp, err := s.sendOTP(primaryTwoFAChannel, purpose2FA, target)
		if err != nil {
			return LoginResult{}, err
		}

		// Adding new otp session to our cache
		s.cache.otp.Add(
			refID,
			&otpSession{
				userID:  user.ID,
				channel: primaryTwoFAChannel,
				purpose: purpose2FA,
				otp:     otp,
			},
			expiresAt,
		)
		return LoginResult{
			Requires2FA: true,
			ReferenceID: refID,
			ExpiresAt:   expiresAt,
		}, nil
	}
	// else if 2fa is not enabled we go on to generate the tokens

	// generate tokens
	jwtID := uuid.New()
	accessToken, err := s.jwt.GenerateToken(
		newAccessToken( // generating a new access token struct and taking its payload version with json compatible tags
			user.ID,
			jwtID,
			s.config.ExpiryTimeAccessToken,
		).Claims(),
	)
	if err != nil {
		return LoginResult{}, fmt.Errorf("service.auth.login - failed to generate jwt: %w", err)
	}
	refreshToken := s.generateRefreshToken()
	// Add a new user session to the database
	err = s.repo.session.Create(ctx, newUserSession(user.ID, jwtID, refreshToken, md))
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ? ----+-----+-----Resend OTP-----+-----+-----

type ResendResult struct {
	ReferenceID string
	ExpiresAt   time.Time
}

// Re sends otp an otp to the according to the provided purpose and channel
func (s *Service) ResendOTP(ctx context.Context, req ResendOTPRequest) (ResendResult, error) {
	if !channelIsValidForOTP(req.Channel) {
		panic("resend otp: invalid channel passed - " + string(req.Channel))
	}
	// Check if the user even exists in our repository
	user, err := s.repo.user.GetByAuthChannel(ctx, req.Channel, req.Value)
	if err != nil {
		if err == repository.ErrNoResults {
			return ResendResult{}, errCommonUserNotFound
		}
		return ResendResult{}, err
	}

	// Check if there is a already active otp session, if so we delete that first to prevent multiple active otp sessions for one user
	var prevRefKey string
	s.cache.otp.LoopFunc(
		// providing a function which receives the key, value and expiresAt when looping over all the items
		func(key string, value *otpSession, _ time.Time) bool {
			if value.userID == user.ID {
				prevRefKey = key
				return true // signals the outer loop to quit
			}
			return false
		},
	)
	if prevRefKey != "" {
		s.cache.otp.Delete(prevRefKey)
	}

	purpose := authPurpose(req.Purpose) // assuming its validated

	// perform logic related to each purpose
	switch purpose {
	// if its for 2fa we change the channel and target to the primary 2fa identifier
	case purpose2FA:
		if user.TwoFAs == nil {
			return ResendResult{}, errResend2FANotEnabled
		}
		req.Channel = user.TwoFAs[0] // considering the first element in the slice to be the primary 2fa identifier
		switch req.Channel {
		case entity.ChannelEmail:
			req.Value = *user.Email
		case entity.ChannelPhone:
			req.Value = *user.Phone
		case entity.ChannelTOTP:
			return ResendResult{}, errResendInvalidIdentifier
		default:
			panic("resend otp: invalid primary channel in repository for user 2fa")
		}
	// if its for reset password we delete any previous reset sessions, same logic as otp sessions above
	case purposeResetPass:
		prevRefKey = ""
		s.cache.resetPassword.LoopFunc(
			func(key string, value resetPasswordSession, expiresAt time.Time) bool {
				if value.userID == user.ID {
					prevRefKey = key
					return true
				}
				return false
			},
		)
		if prevRefKey != "" {
			s.cache.resetPassword.Delete(prevRefKey)
		}
	// if for registration we just check if the channel is a username or not, because that is a invalid channel for registration and should not be sent.
	case purposeRegistration:
		if req.Channel == entity.ChannelUsername {
			return ResendResult{}, errResendInvalidIdentifier
		}
	}

	// Now we send the otp and store it in our cache and return the user a new reference id
	refId := s.generateOTPSessionID() // new otp session id
	expiresAt := time.Now().Add(s.config.ExpiryTimeOTP)
	otp, err := s.sendOTP(req.Channel, purpose, req.Value)
	if err != nil {
		return ResendResult{}, err
	}
	s.cache.otp.Add(
		refId,
		&otpSession{
			userID:  user.ID,
			channel: req.Channel,
			purpose: purpose,
			otp:     otp,
		},
		expiresAt,
	)
	return ResendResult{
		ReferenceID: refId,
		ExpiresAt:   expiresAt,
	}, nil
}

// ? ----+-----+-----Verify otp-----+-----+-----

type VerifyResult struct {
	AccessToken       string
	RefreshToken      string
	IsResetRequest    bool // if this is true the above two values are zeroed out and fields below are populated, else vise versa
	ReferenceID       string
	ExpiresAt         time.Time
	RemainingAttempts int // only populated if returning an errIncorrectOTP error
}

// Verifies the two factor / verification / reset password OTP and generates a token pair / reset pass id for the user.
func (s *Service) VerifyOTP(ctx context.Context, req VerifyOTPRequest, md entity.ClientMetadata) (VerifyResult, error) {
	// retrieve the session from the reference id in the request
	session, expiresAt, ok := s.cache.otp.Get(req.ReferenceID)
	if !ok { // if not found it means either the frontend is trying to reuse the same reference id after expiration or verification
		return VerifyResult{}, errVerifyRefIDNotFound
	}
	if expiresAt.Before(time.Now()) { // if the otp has expired we return a error, the ttl cache automatically removes any values which are expired on a Get call if Cache.Config.LazyDelete is set to true, which is default.
		return VerifyResult{}, errVerifyOTPExpired
	}
	var err error
	if session.channel == entity.ChannelTOTP { // if its a authenticator time based 2 factor code
		session.otp, err = s.otp.GenerateTOTP(session.otp) // in this case secretOTP is actually the secretKey for the TOTP which is used to compute the otp.
		if err != nil {
			return VerifyResult{}, err
		}
	}
	session.attempts++          // increment on every successful attempt
	if session.otp != req.OTP { // we check the otp if it does not match we return early with a incorrect otp error
		remainingAttempts := s.config.AttemptsOTP - session.attempts
		if remainingAttempts <= 0 {
			s.cache.otp.Delete(req.ReferenceID) // delete the session not allowing for any more verification attempts
			return VerifyResult{}, errCommonAttemptsExhausted
		}
		return VerifyResult{RemainingAttempts: remainingAttempts}, errVerifyOTPIncorrect
	}
	// if the otp matches then we remove the reference id from our map immediately
	s.cache.otp.Delete(req.ReferenceID)
	// find the user
	user, err := s.repo.user.GetByID(ctx, session.userID)
	if err != nil {
		if err == repository.ErrNoResults {
			return VerifyResult{}, errCommonUserNotFound // if the user mysteriously got deleted after just trying to log in, register or reset their password...
		}
		return VerifyResult{}, err
	}

	// Switch on the purpose to do purpose related tasks
	switch session.purpose {
	case purposeRegistration: // if registration we need to set user status to verified in the database
		err = s.repo.session.UpdateStatus(ctx, user.ID, entity.StatusVerified)
		if err != nil { // no need to handle for errRepoNoResults because we did that above
			return VerifyResult{}, err
		}
	case purposeResetPass:
		refID := s.generateResetSessionID()
		expiresAt := time.Now().Add(s.config.ExpiryTimeResetPassword)
		s.cache.resetPassword.Add(
			refID,
			resetPasswordSession{
				userID: user.ID,
			},
			expiresAt,
		)
		return VerifyResult{
			IsResetRequest: true,
			ReferenceID:    refID,
			ExpiresAt:      expiresAt,
		}, nil
	}

	// generate new tokens and return them to the user for future usage
	jwtID := uuid.New()
	accessToken, err := s.jwt.GenerateToken(
		newAccessToken(
			user.ID,
			jwtID,
			s.config.ExpiryTimeAccessToken,
		).Claims(),
	)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("service.auth.verify otp - failed to generate jwt: %w", err)
	}
	refreshToken := s.generateRefreshToken()

	// Add a new user session to the database
	err = s.repo.session.Create(ctx, newUserSession(user.ID, jwtID, refreshToken, md))

	if err != nil {
		return VerifyResult{}, err
	}

	return VerifyResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ? ----+-----+-----Forgot password-----+-----+-----

type ForgotPasswordResult struct {
	ReferenceID string
	ExpiresAt   time.Time
}

// Raises a forgot password session request which sends an verification otp to the channel provided and stores a in memory temporary session
func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (ForgotPasswordResult, error) {
	// Find the user in the database
	user, err := s.repo.user.GetByAuthChannel(ctx, req.Channel, req.Value)
	if err != nil {
		if err == repository.ErrNoResults {
			return ForgotPasswordResult{}, errCommonUserNotFound
		}
		return ForgotPasswordResult{}, err // db error
	}

	// Send an otp to the channel and target our user sent us
	otp, err := s.sendOTP(req.Channel, purposeResetPass, req.Value)
	// generate a otp session id and add it to our ttlcache
	refID := s.generateOTPSessionID()
	expiresAt := time.Now().Add(s.config.ExpiryTimeOTP)
	s.cache.otp.Add(
		refID,
		&otpSession{
			userID:  user.ID,
			channel: req.Channel,
			purpose: purposeResetPass,
			otp:     otp,
		},
		expiresAt,
	)
	// send the session id as reference to the user
	return ForgotPasswordResult{
		ReferenceID: refID,
		ExpiresAt:   expiresAt,
	}, nil
}

// ? ----+-----+-----Reset password-----+-----+-----

// updates a user's password
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	// Retrieve the reset session from the cache using the reference id from the request data
	session, expiresAt, ok := s.cache.resetPassword.Get(req.ReferenceID)
	if !ok {
		return errResetSessionNotFound
	}
	// if the reset request expired we return error
	if expiresAt.Before(time.Now()) {
		return errResetSessionExpired
	}

	// else we proceed and update the user's password, we of course hash it.
	err := s.repo.user.UpdatePasswordHash(ctx, session.userID, password.Hash(req.NewPassword))
	if err != nil {
		if err == repository.ErrNoResults {
			return errCommonUserNotFound
		}
		return err
	}
	// delete the session to make sure this reference id cannot be reused
	s.cache.resetPassword.Delete(req.ReferenceID)
	return nil
}

// ? ----+-----+-----Refresh-----+-----+-----

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

// refreshes the token pair with a new pair of tokens
func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (RefreshResult, error) {
	// Do a db lookup with the refresh token's hash
	userSesh, err := s.repo.session.GetByTokenHash(ctx, cryptoutil.GenerateMD5Hash(req.RefreshToken))
	if err != nil {
		if err == repository.ErrNoResults {
			return RefreshResult{}, errRefreshTokenInvalid
		}
		return RefreshResult{}, err // db error
	}
	// Check if the session has expired
	if userSesh.ExpiresAt.Before(time.Now()) {
		// Deleting the user session if it is expired, the user will have to create a new session again by logging in.
		err = s.repo.session.Delete(ctx, "id", userSesh.ID)
		if err != nil {
			return RefreshResult{}, err // db error
		}
		return RefreshResult{}, errRefreshTokenExpired
	}

	// if everything is good we generate both new tokens, we do not have to regenerate and re update the jwt id as its unnecessary.
	// it is a fixed value which is linked with the user session in database
	accessToken, err := s.jwt.GenerateToken(
		newAccessToken(
			userSesh.UserID,
			userSesh.JwtID,
			s.config.ExpiryTimeAccessToken,
		).Claims(),
	)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("service.auth.refresh - failed to generate jwt: %w", err)
	}
	refreshToken := s.generateRefreshToken()

	// update the session with the new refresh token and also update the expires at field to the max capacity again.
	err = s.repo.session.UpdateTokenHashAndExpiry(
		ctx,
		userSesh.ID,
		cryptoutil.GenerateMD5Hash(refreshToken), // we store a hash of the token
		time.Now().Add(s.config.ExpiryTimeRefreshToken),
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
func (s *Service) Logout(ctx context.Context, token AccessToken) error {
	// Remove user session from database
	err := s.repo.session.Delete(ctx, "jwt_id", token.JwtID)
	if err != nil && err != repository.ErrNoResults { // ignore no results
		return err // db error
	}
	// Add the jwt token id to the block list so this gets rejected by the authorization
	s.cache.jwtBlocklist.Add(token.JwtID, struct{}{}, token.ExpiresAt)
	return nil
}

// ? ----+-----+-----Delete Account-----+-----+-----
func (s *Service) DeleteAccount(ctx context.Context, token AccessToken) error {
	err := s.repo.user.Delete(ctx, token.UserID) // assuming there is cascade such that profiles and user sessions also gets deleted.
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
func (s *Service) Add2FA(ctx context.Context, token AccessToken, req Add2FARequest) error {
	// finding the user in the database using the user id from token
	user, err := s.repo.user.GetByID(ctx, token.UserID)
	if err != nil {
		return err
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
	err = s.repo.user.UpdateTwoFAs(ctx, user.ID, user.TwoFAs)
	if err != nil {
		return err
	}
	// returning empty result as there is nothing more we need to signify
	return nil
}

// Removes 2fa for a user
func (s *Service) Remove2FA(ctx context.Context, token AccessToken, req Remove2FARequest) error {
	// finding the user in the database using user id from the token
	user, err := s.repo.user.GetByID(ctx, token.UserID)
	if err != nil {
		return err
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
	err = s.repo.user.UpdateTwoFAs(ctx, user.ID, user.TwoFAs)
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
	ReferenceID string
	TotpUri     string
	ExpiresAt   time.Time
}

// starts a setup for totp
func (s *Service) TotpSetup(ctx context.Context, token AccessToken) (TotpSetupResult, error) {
	// Find user on the database
	user, err := s.repo.user.GetByID(ctx, token.UserID)
	if err != nil {
		return TotpSetupResult{}, err
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
	secretKey, totpUri := s.otp.SetupTOTP(identifier)
	refId := s.generateTotpSessionID()
	expiresAt := time.Now().Add(s.config.TTLCacheCleanIntervalTOTP)
	// adding to cache session
	s.cache.totp.Add(
		refId,
		&totpSession{
			userID:    user.ID,
			secretKey: secretKey,
		},
		expiresAt,
	)
	return TotpSetupResult{
		ReferenceID: refId,
		TotpUri:     totpUri,
		ExpiresAt:   expiresAt,
	}, nil
}

// ? ----+-----+-----Totp Verify-----+-----+-----

type TotpVerifyResult struct {
	RemainingAttempts int
}

// verifies the otp from totp setup and sets secret in database
func (s *Service) TotpVerify(ctx context.Context, token AccessToken, req TotpVerifyRequest) (TotpVerifyResult, error) {
	// Fetching session from cache
	session, expiresAt, ok := s.cache.totp.Get(req.ReferenceID)
	if !ok {
		return TotpVerifyResult{}, errTotpVerifySessionNotFound
	}
	// checking if session and token user's match, otherwise it is a stolen token/reference id
	if session.userID != token.UserID {
		return TotpVerifyResult{}, errCommonUnauthorized
	}
	// checking if session has expired
	if expiresAt.Before(time.Now()) {
		return TotpVerifyResult{}, errTotpVerifySessionExpired
	}
	// calculating the otp
	otp, err := s.otp.GenerateTOTP(session.secretKey)
	if err != nil {
		return TotpVerifyResult{}, err
	}
	session.attempts++ // increment on every attempt
	// comparing against the otp sent
	if req.OTP != otp {
		remainingAttempts := s.config.AttemptsTOTPVerify - session.attempts
		if remainingAttempts <= 0 {
			s.cache.totp.Delete(req.ReferenceID) // delete the session not allowing for any more verification attempts
			return TotpVerifyResult{}, errCommonAttemptsExhausted
		}
		return TotpVerifyResult{remainingAttempts}, errTotpVerifyTOTPIncorrect
	}
	// delete the session if otp matches and verification is complete
	s.cache.totp.Delete(req.ReferenceID)
	// enable totp in database
	err = s.repo.user.UpdateAddTotp(ctx, session.userID, session.secretKey)
	if err != nil {
		return TotpVerifyResult{}, err
	}
	return TotpVerifyResult{}, nil
}
