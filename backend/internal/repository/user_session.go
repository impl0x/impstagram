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

	Create(ctx context.Context, session *entity.UserSession) error

	// ? ----+----+---- Get ----+----+----

	GetByID(ctx context.Context, id uuid.UUID) (*entity.UserSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error)

	// ? ----+----+---- Update ----+----+----

	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus) error
	UpdateTokenHashAndExpiry(ctx context.Context, id uuid.UUID, tokenHash string, expiry time.Time) error

	// ? ----+----+---- Delete ----+----+----

	Delete(ctx context.Context, column string, value any) error
}
