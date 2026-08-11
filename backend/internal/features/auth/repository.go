package auth

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, userID uuid.UUID) (*User, error)
	FindByChannel(ctx context.Context, channel authChannel, identifierValue string) (*User, error)
	Create(ctx context.Context, user *User) error

	Get(ctx context.Context, id uuid.UUID) (*User, error)
}
