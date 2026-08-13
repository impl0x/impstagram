package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	// users table
	FindUserByID(ctx context.Context, userID uuid.UUID) (*user, error)
	FindUserByChannel(ctx context.Context, channel authChannel, target string) (*user, error)
	CreateUser(ctx context.Context, user *user) error

	// user_sessions table
	FindSessionByToken(ctx context.Context, tokenHash string) (*userSession, error)
	CreateSession(ctx context.Context, session *userSession) error
	UpdateSessionToken(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
}
