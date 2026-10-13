package repository

import (
	"backend/internal/entity"
	"context"
	"uuid"
)

// contains profile related methods
type ProfileRepository interface {

	// ? -----+-----+----- Create -----+-----+-----

	// Creates a new profile in the repository
	Create(ctx context.Context, profile *entity.Profile) error

	// ? -----+-----+----- Get -----+-----+-----

	// Gets a profile using UserID
	GetByUserID(ctx context.Context, id uuid.UUID) (*entity.Profile, error)
	// Gets a profile using Username
	GetByUsername(ctx context.Context, username string) (*entity.Profile, error)

	// ? -----+-----+----- Extra -----+-----+-----

	// Checks if a profile with the username is already present
	CheckUsername(ctx context.Context, username string) (bool, error)
}
