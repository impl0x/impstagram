package repository

import (
	"backend/internal/entity"
	"context"
	"uuid"
)

// Contains table user_sessions
type UserSessionRepository interface {
	Create(ctx context.Context, session *entity.UserSession) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.UserSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error)
	Update(ctx context.Context, id uuid.UUID, column string, value any) error
	Delete(ctx context.Context, column string, value any) error
}
