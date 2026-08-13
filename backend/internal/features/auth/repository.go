package auth

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// users table
	FindUserByID(ctx context.Context, userID uuid.UUID) (*user, error)
	FindUserByChannel(ctx context.Context, channel authChannel, target string) (*user, error)
	CreateUser(ctx context.Context, user *user) error

	// user_sessions table
	FindSessionByToken(ctx context.Context, tokenHash string)
	FindSessionByUserID(ctx context.Context, userID uuid.UUID)
	CreateSession(ctx context.Context, session *userSession) error
}
