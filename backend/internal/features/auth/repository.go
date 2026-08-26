package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type repository interface {
	// users
	findUserByID(ctx context.Context, userID uuid.UUID) (*user, error)
	findUserByChannel(ctx context.Context, channel authChannel, target string) (*user, error)
	createUser(ctx context.Context, user *user) (uuid.UUID, error)
	updateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	updateUserStatus(ctx context.Context, userID uuid.UUID, status accountStatus) error
	updateUser2FA(ctx context.Context, userID uuid.UUID, twoFAs twoFAs) error

	// user_sessions
	findSessionByToken(ctx context.Context, tokenHash string) (*userSession, error)
	createSession(ctx context.Context, session *userSession) error
	updateSessionToken(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error
	deleteSessionByID(ctx context.Context, id uuid.UUID) error
	deleteSessionByJwtID(ctx context.Context, jwtID uuid.UUID) error
}

// Repository errors
var (
	errRepoNoResults      = errors.New("no results found")
	errRepoTooManyResults = errors.New("too many results")
)
