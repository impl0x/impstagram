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
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Temp2FAState struct {
	UserID    uuid.UUID
	SecretOTP string
	ExpiresAt time.Time
}

type Service struct {
	Jwt   jwt.Jwt
	Email email.Client
	Repo  Repository

	Pending2FA map[string]Temp2FAState // string is the session uuid
}

var ErrAlreadyExistingUser = errors.New("User already exists")
var ErrNotOldEnough = errors.New("User not old enough, must be minimum of " + strconv.Itoa(int(config.MinAge)) + " years old to use this service")
var ErrTooOld = errors.New("User too old, cannot create account")

func (s *Service) Register(ctx context.Context, req RegisterRequest) (string, error) {

	userDob, err := dob.NewDobFromString(req.Dob)

	if err != nil {
		return "", err
	}

	userAge := userDob.Age()

	if userAge < config.MinAge {
		return "", ErrNotOldEnough
	} else if userAge > config.MaxAge {
		return "", ErrTooOld
	}

	var user *User

	if req.Email != "" {
		user, err = s.Repo.FindByEmail(ctx, req.Email)
	} else if req.Phone != "" {
		user, err = s.Repo.FindByPhone(ctx, req.Phone)
	} else {
		return "", ErrMissingIdentifier
	}

	if err != nil {
		return "", err
	}

	if user != nil { // if db returned a user
		return "", ErrAlreadyExistingUser // ignoring database level errors as of now.
	}
	// todo: have email verification here
	passwordHash := password.Hash(req.Password)
	token := s.Jwt.GenerateToken()

	s.Repo.Create(ctx, NewUser(req, passwordHash))

	return token, nil
}

type LoginResult struct {
	Token           string
	Requires2FA     bool   // if this is false then below all fields are zero'd out, else the above token is zero value'd
	ReferenceId     string // Used to link the upcoming OTP request
	TwoFAIdentifier Identifier
}

// var ErrMissingIdentifier = errors.New("Need at least one of the following: email, phone or username")
// var ErrUserNotFound = errors.New("User not found")
// var ErrInvalidCredentials = errors.New("Invalid credentials")
// var ErrUserBanned = errors.New("User is banned from accessing this service")
// var ErrUserUnverified = errors.New("User is unverified, please verify with your email/phone first")d

var (
	ErrMissingIdentifier = errors.New("missing identifier")
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrUserBanned        = errors.New("user banned")
	ErrUserUnverified    = errors.New("user unverified")
)

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResult, error) {
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
		return LoginResult{}, ErrMissingIdentifier
	}

	if err != nil { // database error
		return LoginResult{}, err
	}

	if user == nil {
		return LoginResult{}, ErrUserNotFound
	}

	if !password.Compare(req.Password, user.PasswordHash) {
		return LoginResult{}, ErrIncorrectPassword
	}

	if user.Status == StatusBanned {
		return LoginResult{}, ErrUserBanned
	} else if user.Status == StatusUnverified {
		return LoginResult{}, ErrUserUnverified
	}

	// If user has 2Fa enabled ask for otp.
	if user.TwoFA != nil {
		twoFAOTP, err := token.GenerateOTP()
		if err != nil { // otp generation error
			return LoginResult{}, err
		}
		err = nil
		TwoFAIdentifier := user.TwoFA[0] // by default the first element is the priority 2FA identifier
		switch TwoFAIdentifier {
		case IdentifierEmail:
			err = s.Email.Send(email.NewSendRequest(user.Email, email.SubjectTwoFa, email.HtmlOtp.Format(twoFAOTP)))
		case IdentifierPhone:
			//todo: send sms with 2fa code
		case IdentifierTOTP:
			// todo: setup totp
		default:
			// impossible case so throw panic
			panic("dev fucked up somewhere")
		}
		if err != nil { // send otp error
			return LoginResult{}, err
		}
		referenceId := token.GenerateReferenceID()
		s.Pending2FA[referenceId] = Temp2FAState{user.Id, twoFAOTP, time.Now().Add(config.OTPExpiryTime)}
		return LoginResult{
			Requires2FA:     true,
			ReferenceId:     referenceId,
			TwoFAIdentifier: TwoFAIdentifier, // by default we use the first priority 2FA identifier
		}, nil
	}

	token := s.Jwt.GenerateToken() // jwt is still a todo
	return LoginResult{Token: token}, nil
}
