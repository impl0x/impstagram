package repository

import (
	"backend/internal/entity"
	"context"
	"time"
	"uuid"
)

// Contains user sessions related methods
type UserSessionRepository interface {
	// ? ----+----+---- Create ----+----+----

	// Creates a new user session
	Create(ctx context.Context, session *entity.UserSession) error

	// ? ----+----+---- Get ----+----+----

	// Gets a user session by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.UserSession, error)
	// Gets a user session by its token hash field
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error)

	// ? ----+----+---- Update ----+----+----

	// Updates the token hash and expiry time of session by its ID
	UpdateTokenHashAndExpiry(ctx context.Context, id uuid.UUID, tokenHash string, expiry time.Time) error

	// ? ----+----+---- Delete ----+----+----

	// Deletes a user session by its ID
	DeleteByID(ctx context.Context, id uuid.UUID) error
	// Deletes a user session by its jwt id field
	DeleteByJwtID(ctx context.Context, jwtID uuid.UUID) error
}
