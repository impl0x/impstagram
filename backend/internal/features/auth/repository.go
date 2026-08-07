package auth

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)

	Create(ctx context.Context, user *User) error

	Get(ctx context.Context, id uuid.UUID) (*User, error)
}
