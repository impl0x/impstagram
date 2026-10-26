package postgres

import (
	"backend/internal/database"
	"backend/internal/entity"
	"backend/pkg/postgres"
	"context"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

// INFO
// Contains profiles table methods

type Profiles struct {
	postgres.Postgres
}

// ? -----+-----+----- Create -----+-----+-----

func (pg Profiles) Create(ctx context.Context, profile *entity.Profile) error {
	cols := make([]string, 0, 7)
	vals := make([]any, 0, 7)
	totalLen := 43 + len(database.TableProfiles)
	if profile.DisplayName != nil {
		cols = append(cols, "display_name")
		vals = append(vals, profile.DisplayName)
		totalLen += 16
	}
	if profile.AvatarUrl != nil {
		cols = append(cols, "avatar_url")
		vals = append(vals, profile.AvatarUrl)
		totalLen += 14
	}
	if profile.Bio != nil {
		cols = append(cols, "bio")
		vals = append(vals, profile.Bio)
		totalLen += 7
	}
	var t time.Time
	if profile.UpdatedAt != t {
		pg.Log.Warn("DB: Providing updated_at in create profile is not good, let the database create that by default for you")
		cols = append(cols, "updated_at")
		vals = append(vals, profile.UpdatedAt)
		totalLen += 14
	}
	if !profile.IsPrivate {
		cols = append(cols, "is_private")
		vals = append(vals, profile.IsPrivate)
		totalLen += 14
	}
	// ignore created_at even if provided. hardcoded to not set it via service call
	cols = append(cols, "user_id", "username")
	vals = append(vals, profile.UserID, profile.Username)

	var buf strings.Builder
	buf.Grow(totalLen)
	buf.WriteString("INSERT INTO " + database.TableUsers + " (")
	colWriter(&buf, cols)
	buf.WriteString(") VALUES (")
	valWriter(&buf, len(vals))
	buf.WriteString(")")

	_, err := pg.Pool.Exec(ctx, buf.String(), vals...)
	if err != nil {
		handlePgxError(fmt.Errorf("profiles create, %w", err))
	}
	return nil
}

// ? -----+-----+----- Get -----+-----+-----

func (pg Profiles) get(ctx context.Context, by string, value any) (*entity.Profile, error) {
	var buf strings.Builder
	buf.Grow(97 + len(database.TableProfiles) + len(by))
	buf.WriteString("SELECT user_id, username, display_name, avatar_url, is_private, bio, updated_at FROM " + database.TableUsers + " WHERE ")
	buf.WriteString(by)
	buf.WriteString(" = $1")
	var profile entity.Profile
	err := pg.Pool.QueryRow(ctx, buf.String(), value).
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
		return nil, handlePgxError(fmt.Errorf("profiles get by %s, %w", by, err))
	}
	return &profile, nil
}

func (pg Profiles) GetByUserID(ctx context.Context, id uuid.UUID) (*entity.Profile, error) {
	return pg.get(ctx, "user_id", id)
}
func (pg Profiles) GetByUsername(ctx context.Context, username string) (*entity.Profile, error) {
	return pg.get(ctx, "username", username)
}

// ? -----+-----+----- Update -----+-----+-----

func (pg Profiles) update(ctx context.Context, id uuid.UUID, col string, val any) error {
	var buf strings.Builder
	buf.Grow(48 + len(database.TableProfiles) + len(col))
	buf.WriteString("UPDATE " + database.TableProfiles + " SET ")
	buf.WriteString(col)
	buf.WriteString(" = $1, updated_at = $2 WHERE id = $3")

	cmdTag, err := pg.Pool.Exec(ctx, buf.String(), val, time.Now(), id)
	if err != nil {
		return handlePgxError(fmt.Errorf("user sessions update col %s, %w", col, err))
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}

func (pg Profiles) UpdateUsername(ctx context.Context, id uuid.UUID, username string) error {
	return pg.update(ctx, id, "username", username)
}

func (pg Profiles) UpdateDisplayName(ctx context.Context, id uuid.UUID, displayName string) error {
	return pg.update(ctx, id, "display_name", displayName)
}

func (pg Profiles) UpdateAvatarUrl(ctx context.Context, id uuid.UUID, url string) error {
	return pg.update(ctx, id, "avatar_url", url)
}

func (pg Profiles) UpdateIsPrivate(ctx context.Context, id uuid.UUID, isPrivate bool) error {
	return pg.update(ctx, id, "is_private", isPrivate)
}

func (pg Profiles) UpdateBio(ctx context.Context, id uuid.UUID, bio string) error {
	return pg.update(ctx, id, "bio", bio)
}

// ? -----+-----+----- Extra -----+-----+-----

func (pg Profiles) CheckUsername(ctx context.Context, username string) (bool, error) {
	var id uuid.UUID
	err := pg.Pool.QueryRow(ctx, "SELECT user_id FROM "+database.TableProfiles+" WHERE username = $1", username).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("profiles check username, %w", err)
	}
	return true, nil
}
