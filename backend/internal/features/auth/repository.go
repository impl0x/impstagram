package auth

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, userID uuid.UUID) (*User, error)
	FindByChannel(ctx context.Context, channel authChannel, target string) (*User, error)
	Create(ctx context.Context, user *User) error	
}
