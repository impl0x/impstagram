package repository

import (
	"backend/internal/entity"
	"context"
	"time"
	"uuid"
)

// Contains table user_sessions
type UserSessionRepository interface {
	// ? ----+----+---- Create ----+----+----

	Create(ctx context.Context, session *entity.UserSession) (uuid.UUID, error)

	// ? ----+----+---- Get ----+----+----

	GetByID(ctx context.Context, id uuid.UUID) (*entity.UserSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error)

	// ? ----+----+---- Update ----+----+----

	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus) error
	UpdateTokenAndExpiry(ctx context.Context, id uuid.UUID, token string, expiry time.Time) error

	// ? ----+----+---- Delete ----+----+----

	Delete(ctx context.Context, column string, value any) error
}
