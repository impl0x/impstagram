package auth

import (
	"backend/internal/config"
	"backend/internal/pkg/dob"
	"backend/internal/pkg/email"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/otp"
	"backend/internal/pkg/password"
	"context"
	"errors"
	"strconv"
)

type Service struct {
	jwt   jwt.Jwt
	email email.Client
	repo  Repository
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
		user, err = s.repo.FindByEmail(ctx, req.Email)
	} else if req.Phone != "" {
		user, err = s.repo.FindByPhone(ctx, req.Phone)
	} else {
		return "", ErrMissingIdentifier
	}

	if err != nil {
		return "", err
	}

	if user != nil { // if db returned a user
		return "", ErrAlreadyExistingUser // ignoring database level errors as of now.
	}

	passwordHash := password.Hash(req.Password)
	token := s.jwt.GenerateToken()

	s.repo.Create(ctx, NewUser(req, passwordHash))

	return token, nil
}

var ErrMissingIdentifier = errors.New("Need at least one of the following: email, phone or username")
var ErrInvalidCredentials = errors.New("Invalid credentials")
var ErrUserBanned = errors.New("User is banned from accessing this service")
var ErrUserUnverified = errors.New("User is unverified, please verify with your email/phone first")

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	var user *User
	var err error

	switch {
	case req.Username != "":
		user, err = s.repo.FindByUsername(ctx, req.Username)
	case req.Email != "":
		user, err = s.repo.FindByEmail(ctx, req.Email)
	case req.Phone != "":
		user, err = s.repo.FindByPhone(ctx, req.Phone)
	default:
		return "", ErrMissingIdentifier
	}

	if err != nil {
		return "", err
	}

	if user == nil {
		return "", ErrInvalidCredentials
	}

	if !password.Compare(req.Password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	if user.Status == StatusBanned {
		return "", ErrUserBanned
	} else if user.Status == StatusUnverified {
		return "", ErrUserUnverified
	}

	err = nil
	if user.TwoFA != nil {
		twoFaOTP, err := otp.GenerateOTP()
		if err != nil {
			return "", err
		}
		switch user.TwoFA[0] {
		case IdentifierEmail:
			// send a 2fa email
			err = s.email.Send(email.NewSendRequest(user.Email, email.SubjectTwoFa, email.HtmlOtp.Format(twoFaOTP)))
		case IdentifierPhone:
			// send sms with 2fa code
		default:
			// impossible case so throw panic
			panic("dev fucked up somewhere")
		}
		if err != nil {
			return "", err
		}
	}

	token := s.jwt.GenerateToken()
	return token, nil
}
