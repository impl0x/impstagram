package auth

import (
	"backend/internal/config"
	"backend/internal/pkg/dob"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/password"
	"context"
	"errors"
	"strconv"
)

type Service struct {
	jwt  jwt.Jwt
	repo Repository
}

var ErrAlreadyExistingUser = errors.New("User already exists")
var ErrNotOldEnough = errors.New("User not old enough, must be minimum of " + strconv.Itoa(int(config.MinAge)) + " years old to use this service")

func (s *Service) Register(ctx context.Context, req RegisterRequest) (string, error) {

	userDob, err := dob.NewDobFromString(req.Dob)

	if err != nil {
		return "", err
	}

	if !dob.IsOldEnough(config.MinAge, userDob) {
		return "", ErrNotOldEnough
	}

	var user User
	println(user)//to be removed

	if req.Username != "" {
		user, err = s.repo.FindByUsername(ctx, req.Username)
	} else if req.Email != "" {
		user, err = s.repo.FindByEmail(ctx, req.Email)
	} else if req.Phone != "" {
		user, err = s.repo.FindByPhone(ctx, req.Phone)
	} else {
		return "", ErrMissingIdentifier
	}

	if err != nil {
		return "", ErrInvalidCredentials // ignoring database level errors as of now.
	}

	passwordHash := password.Hash(req.Password)
	token := s.jwt.GenerateToken()

	s.repo.Create(ctx, NewUser(req, passwordHash))

	return token, nil
}

var ErrMissingIdentifier = errors.New("Need at least one of the following: email, phone or username")
var ErrInvalidCredentials = errors.New("Invalid credentials")

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	var user User
	var err error

	if req.Username != "" {
		user, err = s.repo.FindByUsername(ctx, req.Username)
	} else if req.Email != "" {
		user, err = s.repo.FindByEmail(ctx, req.Email)
	} else if req.Phone != "" {
		user, err = s.repo.FindByPhone(ctx, req.Phone)
	} else {
		return "", ErrMissingIdentifier
	}

	if err != nil {
		return "", ErrInvalidCredentials // ignoring database level errors as of now.
	}

	if !password.Compare(req.Password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}
	token := s.jwt.GenerateToken()
	return token, nil
}
