package auth

import (
	"context"
	"errors"
	"time"
	"uuid"
)

type tableName = string

// database table names
const (
	tableUsers        tableName = "users"
	tableUserSessions tableName = "user_sessions"
	tableProfiles     tableName = "profiles"
)

// ? INFO:
// Repository containing the data about the entities
// ! Ownership and usage:
// owned by itself
// used by service

type repository interface {
	// users
	findUserByID(ctx context.Context, userID uuid.UUID) (*userModel, error)
	findUserByChannel(ctx context.Context, channel authChannel, target string) (*userModel, error)
	createUser(ctx context.Context, user *userModel) (uuid.UUID, error)
	updateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	updateUserStatus(ctx context.Context, userID uuid.UUID, status accountStatus) error
	updateUser2FA(ctx context.Context, userID uuid.UUID, twoFAs twoFAs) error
	enableTotp(ctx context.Context, userID uuid.UUID, totpKey string) error

	// user_sessions
	findSessionByToken(ctx context.Context, tokenHash string) (*userSessionModel, error)
	createSession(ctx context.Context, session *userSessionModel) error
	updateSessionToken(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error
	deleteSessionByID(ctx context.Context, id uuid.UUID) error
	deleteSessionByJwtID(ctx context.Context, jwtID uuid.UUID) error
}

// Repository errors
var (
	errRepoNoResults      = errors.New("no results found")
	errRepoTooManyResults = errors.New("too many results")
)
