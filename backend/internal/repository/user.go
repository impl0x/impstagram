package repository

import (
	"backend/internal/entity"
	"context"
	"uuid"
)

// Contains user related methods
type UserRepository interface {
	// ? -----+-----+----- Create -----+-----+-----

	// Creates new user to both users table and profiles table, uses transactions to prevent dirty state
	Create(ctx context.Context, user *entity.User, username string) (uuid.UUID, error)

	// ? -----+-----+----- Get -----+-----+-----

	// Gets a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// Gets a user by [entity.AuthChannel], must be a valid channel for fetching users otherwise *panics*.
	GetByAuthChannel(ctx context.Context, channel entity.AuthChannel, value string) (*entity.User, error)

	// ? -----+-----+----- Update -----+-----+-----

	// Updates the status field for a user by its ID
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus) error
	// Updates the password hash field for a user by its ID
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error
	// Updates the 2fa slice for a user by its ID
	UpdateTwoFAs(ctx context.Context, id uuid.UUID, twoFAs []entity.AuthChannel) error
	// Updates the totp secret key field and appends [entity.ChannelTOTP] in 2fa slice
	UpdateAddTotp(ctx context.Context, id uuid.UUID, secretKey string) error

	// ? -----+-----+----- Delete -----+-----+-----

	// Deletes a user by ID
	Delete(ctx context.Context, id uuid.UUID) error
}
