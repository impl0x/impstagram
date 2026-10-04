package repository

import (
	"backend/internal/entity"
	"context"
	"uuid"
)

// Contains tables users
type UserRepository interface {
	// ? ----+----+---- Create ----+----+----

	// Inserts new user to both users table and profiles table, uses transactions to prevent dirty state
	Create(ctx context.Context, user *entity.User, username string) (uuid.UUID, error)

	// ? ----+----+---- Get ----+----+----

	// Finds a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// Finds a user by [entity.AuthChannel], must be a valid channel for fetching users otherwise *panics*.
	GetByAuthChannel(ctx context.Context, channel entity.AuthChannel, value string) (*entity.User, error)

	// ? ----+----+---- Update ----+----+----

	UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error
	UpdateTwoFAs(ctx context.Context, id uuid.UUID, twoFAs []entity.AuthChannel) error
	UpdateAddTotp(ctx context.Context, id uuid.UUID, secretKey string) error

	// ? ----+----+---- Delete ----+----+----

	Delete(ctx context.Context, id uuid.UUID) error
}
