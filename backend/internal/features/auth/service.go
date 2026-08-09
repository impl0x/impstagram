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
	Jwt   jwt.Jwt
	Email email.Client
	Repo  Repository

	mu2FA      sync.RWMutex
	Pending2FA map[string]temp2FAState // string is the session uuid
}

type temp2FAState struct {
	userID    uuid.UUID
	secretOTP string
	expiresAt time.Time
}

// ? 2fa helper functions

// The reason these are all separate functions is to ensure mutexes are used and its separated from the business logic,
// it is a bit redundant to do this but I feel its better than putting mutexes in the business logic every time

func (s *Service) addNewPending2FA(refID string, identifier identifier, userID uuid.UUID, otp string) {
	var expiryTime time.Duration
	switch identifier {
	case identifierEmail, identifierPhone:
		expiryTime = config.OTPExpiryTime
	case identifierTOTP:
		expiryTime = config.TOTPExpiryTime
	default:
		panic("invalid identifier, please update code if added new identifiers")
	}
	twoFAState := temp2FAState{userID, otp, time.Now().Add(expiryTime)}
	s.mu2FA.Lock()
	defer s.mu2FA.Unlock()
	s.Pending2FA[refID] = twoFAState
}

func (s *Service) getPending2FA(refID string) (state temp2FAState, ok bool) {
	s.mu2FA.RLock()
	defer s.mu2FA.RUnlock()
	state, ok = s.Pending2FA[refID]
	return
}

func (s *Service) removePending2FA(refID string) {
	s.mu2FA.Lock()
	defer s.mu2FA.Unlock()
	delete(s.Pending2FA, refID)
}

// ? Register

type registerResult struct {
	referenceId     string // Used to link the upcoming OTP request
	twoFAIdentifier identifier
}

var (
	errAlreadyExistingUser = errors.New("user already exists")
	errNotOldEnough        = errors.New("user not old enough")
	errTooOld              = errors.New("user too old")
)

func (s *Service) register(ctx context.Context, req registerRequest) (registerResult, error) {

	userDob, err := dob.NewDobFromString(req.Dob)

	if err != nil {
		return registerResult{}, err // returns dob sentinel errors, which we handle in the handler.handleError function
	}

	userAge := userDob.Age()

	if userAge < config.MinAge {
		return registerResult{}, errNotOldEnough
	} else if userAge > config.MaxAge {
		return registerResult{}, errTooOld
	}

	var user *User
	var primaryIdentifier identifier

	if req.Email != "" {
		user, err = s.Repo.FindByEmail(ctx, req.Email)
		primaryIdentifier = identifierEmail
	} else if req.Phone != "" {
		user, err = s.Repo.FindByPhone(ctx, req.Phone)
		primaryIdentifier = identifierPhone
	} else {
		return registerResult{}, errMissingIdentifier
	}

	if err != nil {
		return registerResult{}, err // ignoring database level errors as of now.
	}

	if user != nil { // if db returned a user
		return registerResult{}, errAlreadyExistingUser // yes i acknowledge that user has chances of being banned/unverified, but this is intended. We want user to login and then hit those errors if they exist.
	}

	passwordHash := password.Hash(req.Password)

	user = NewUser(req, passwordHash) // returns a unverified user by default
	err = s.Repo.Create(ctx, user)    // we create a user before sending otp
	if err != nil {
		return registerResult{}, nil
	}

	verificationOTP, err := token.GenerateOTP()
	if err != nil {
		return registerResult{}, nil
	}
	switch primaryIdentifier {
	case identifierEmail:
		err = s.Email.Send(email.NewSendRequest(user.Email, email.SubjectVerifyEmail, email.HtmlOtp.Format(verificationOTP)))
	case identifierPhone:
		// todo: send otp on phone
	}
	if err != nil { // send otp error
		return registerResult{}, err
	}

	referenceId := token.GenerateReferenceID()
	s.addNewPending2FA(referenceId, primaryIdentifier, user.ID, verificationOTP)

	return registerResult{
		referenceId:     referenceId,
		twoFAIdentifier: primaryIdentifier,
	}, nil
} // one of the design choices in this function was to first create the user in the database and then send a verification email/sms
// this is because if the db call fails then we exit early with no otp sent,
// and if the db call succeeds but sending email fails we have a unverified user but no otp
// if the frontend shows the internal error the user may try to register again which will trigger a already exists
// which then will prompt the user to login and if they login they will be prompted to verify their email which is a secure workflow,
// although more work for the user but this is considering that this is the worst case scenario.
// better than a dangling otp with no registered user.

// ? Login

type loginResult struct {
	token           string
	requires2FA     bool   // if this is false then below all fields are zero'd out, else the above token is zero value'd
	referenceID     string // Used to link the upcoming OTP request
	twoFAIdentifier identifier
}

var (
	errMissingIdentifier = errors.New("missing identifier")
	errUserNotFound      = errors.New("user not found")
	errIncorrectPassword = errors.New("incorrect password")
	errUserBanned        = errors.New("user banned")
	errUserUnverified    = errors.New("user unverified")
)

func (s *Service) login(ctx context.Context, req loginRequest) (loginResult, error) {
	var user *User
	var err error

	switch {
	case req.Username != "":
		user, err = s.Repo.FindByUsername(ctx, req.Username)
	case req.Email != "":
		user, err = s.Repo.FindByEmail(ctx, req.Email)
	case req.Phone != "":
		user, err = s.Repo.FindByPhone(ctx, req.Phone)
	default:
		return loginResult{}, errMissingIdentifier
	}

	if err != nil { // database error
		return loginResult{}, err
	}

	if user == nil {
		return loginResult{}, errUserNotFound
	}

	if !password.Compare(req.Password, user.PasswordHash) {
		return loginResult{}, errIncorrectPassword
	}

	if user.Status == statusBanned {
		return loginResult{}, errUserBanned
	} else if user.Status == statusUnverified {
		return loginResult{}, errUserUnverified
	}

	// If user has 2Fa enabled ask for otp.
	if user.TwoFA != nil {
		twoFAOTP, err := token.GenerateOTP()
		if err != nil { // otp generation error
			return loginResult{}, err
		}
		primaryIdentifier := user.TwoFA[0] // by default the first element is the primary 2FA identifier
		switch primaryIdentifier {
		case identifierEmail:
			err = s.Email.Send(email.NewSendRequest(user.Email, email.SubjectTwoFa, email.HtmlOtp.Format(twoFAOTP)))
		case identifierPhone:
			//todo: send sms with 2fa code
		case identifierTOTP:
			// todo: setup totp
		default:
			// impossible case so throw panic
			panic("invalid identifier, please update code if added new identifiers")
		}
		if err != nil { // send otp error
			return loginResult{}, err
		}
		referenceId := token.GenerateReferenceID()
		s.addNewPending2FA(referenceId, primaryIdentifier, user.ID, twoFAOTP)
		return loginResult{
			requires2FA:     true,
			referenceID:     referenceId,
			twoFAIdentifier: primaryIdentifier, // by default we use the first priority 2FA identifier
		}, nil
	}

	token := s.Jwt.GenerateToken() // todo: jwt is still a todo
	return loginResult{token: token}, nil
}

// ? verify otp

type verifyResult struct {
	token string
}

var errRefIDNotFound = errors.New("reference ID not found")

func (s *Service) verify(ctx context.Context, req verifyRequest) (verifyResult, error) { // token, error
	state, ok := s.getPending2FA(req.ReferenceID)
	if !ok {
		return verifyResult{}, errRefIDNotFound
	}

}
