package repository

import (
	"backend/internal/entity"
	"context"
	"uuid"
)

// Contains tables users, user_sessions and profiles
type UserRepository interface {
	// Inserts new user to both users table and profiles table, uses transactions to prevent dirty state
	Create(ctx context.Context, user *entity.User, username string) (uuid.UUID, error)
	// Finds a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// Finds a user by [entity.AuthChannel], must be a valid channel for fetching users otherwise *panics*.
	GetByAuthChannel(ctx context.Context, channel entity.AuthChannel, value string) (*entity.User, error)
	Update(ctx context.Context, id uuid.UUID, column string, value any) error
	Delete(ctx context.Context, id uuid.UUID) error
}
