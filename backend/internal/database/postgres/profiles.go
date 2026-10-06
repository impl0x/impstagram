package postgres

import (
	"backend/internal/database"
	"backend/internal/entity"
	"backend/pkg/postgres"
	"context"
	"fmt"
	"uuid"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

// INFO
// Contains profiles table methods

type Profiles struct {
	postgres.Postgres
}

func NewProfiles(pg postgres.Postgres) Profiles {
	return Profiles{pg}
}

// ? -----+-----+----- Create -----+-----+-----

func (pg Profiles) Create(ctx context.Context, profile *entity.Profile) error {
	if profile == nil {
		panic("profile is nil")
	}
	ag := newArgBuilder()
	if profile.DisplayName != nil {
		ag.add("display_name", profile.DisplayName)
	}
	if profile.AvatarUrl != nil {
		ag.add("avatar_url", profile.AvatarUrl)
	}
	if profile.Bio != nil {
		ag.add("bio", profile.Bio)
	}
	// ignore created_at even if provided. hardcoded to not set it via service call
	ag.addCols("user_id", "username", "is_private", "updated_at")
	ag.addVals(profile.UserID, profile.Username, profile.IsPrivate, profile.UpdatedAt)
	sql, args, err := pg.Builder.
		Insert(ag.unwrapCols()).
		Into(database.TableProfiles).
		Values(ag.getVals()...).
		ToSql()
	if err != nil {
		return fmt.Errorf("ProfilesRepo - create - pg.Builder: %w", err)
	}
	_, err = pg.Pool.Exec(ctx, sql, args...)
	if err != nil {
		handlePgxError(err)
	}
	return nil
}

// ? -----+-----+----- Get -----+-----+-----

func (pg Profiles) get(ctx context.Context, by string, value any) (*entity.Profile, error) {
	sql, args, err := pg.Builder.
		Select("user_id, username, display_name, avatar_url, is_private, bio, updated_at").
		From(database.TableProfiles).
		Where(squirrel.Eq{by: value}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("ProfilesRepo - get - pg.Builder: %w", err)
	}
	var profile entity.Profile
	err = pg.Pool.QueryRow(ctx, sql, args...).
		Scan(
			&profile.UserID,
			&profile.Username,
			&profile.DisplayName,
			&profile.AvatarUrl,
			&profile.IsPrivate,
			&profile.Bio,
			&profile.UpdatedAt,
		)
	if err != nil {
		return nil, handlePgxError(err)
	}
	return &profile, nil
}

func (pg Profiles) GetByUserID(ctx context.Context, id uuid.UUID) (*entity.Profile, error) {
	return pg.get(ctx, "user_id", id)
}
func (pg Profiles) GetByUsername(ctx context.Context, username string) (*entity.Profile, error) {
	return pg.get(ctx, "username", username)
}

// ? -----+-----+----- Extra -----+-----+-----

func (pg Profiles) CheckUsername(ctx context.Context, username string) (bool, error) {
	sql, args, err := pg.Builder.
		Select("user_id").
		From(database.TableProfiles).
		Where(squirrel.Eq{"username": username}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("ProfilesRepo - check username - pg.Builder: %w", err)
	}
	var id uuid.UUID
	err = pg.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
