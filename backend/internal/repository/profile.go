package repository

import (
	"backend/internal/entity"
	"context"
	"uuid"
)

type ProfileRepository interface {

	// ? -----+-----+----- Create -----+-----+-----

	Create(ctx context.Context, profile *entity.Profile) error

	// ? -----+-----+----- Get -----+-----+-----

	GetByUserID(ctx context.Context, id uuid.UUID) (*entity.Profile, error)
	GetByUsername(ctx context.Context, username string) (*entity.Profile, error)

	// ? -----+-----+----- Extra -----+-----+-----

	CheckUsername(ctx context.Context, username string) (bool, error)
}
