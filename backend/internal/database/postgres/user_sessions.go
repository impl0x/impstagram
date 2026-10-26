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
// Contains user_sessions table methods

type UserSessions struct {
	postgres.Postgres
}

// ? -----+-----+----- Create -----+-----+-----

func (pg UserSessions) Create(ctx context.Context, session *entity.UserSession) error {
	totalLen := 72 + len(database.TableUserSessions)
	cols := make([]string, 0, 9)
	vals := make([]any, 0, 9)
	if session.ID != uuid.Nil() {
		cols = append(cols, "id")
		vals = append(vals, session.ID)
		totalLen += 6
		pg.Log.Warn("DB: Providing ID in create user session is not good, let the database create that by default for you.")
	}
	if session.OSName != nil {
		cols = append(cols, "os_name")
		vals = append(vals, session.OSName)
		totalLen += 11
	}
	if session.BrowserName != nil {
		cols = append(cols, "browser_name")
		vals = append(vals, session.BrowserName)
		totalLen += 16
	}
	if session.DeviceType != nil {
		cols = append(cols, "device_type")
		vals = append(vals, session.DeviceType)
		totalLen += 15
	}
	var t time.Time
	if session.CreatedAt != t {
		pg.Log.Warn("DB: created_at provided in user sessions create, will not be set and is not recommended to set")
	}
	cols = append(cols, "jwt_id", "user_id", "token_hash", "expires_at")
	vals = append(vals, session.JwtID, session.UserID, session.TokenHash, session.ExpiresAt)

	var buf strings.Builder
	buf.Grow(totalLen)
	buf.WriteString("INSERT INTO " + database.TableUserSessions + " (")
	colWriter(&buf, cols)
	buf.WriteString(") VALUES (")
	valWriter(&buf, len(vals))
	buf.WriteByte(')')

	_, err := pg.Pool.Exec(ctx, buf.String(), vals...)
	if err != nil {
		handlePgxError(fmt.Errorf("user sessions create, %w", err))
	}
	return nil
}

// ? -----+-----+----- Get -----+-----+-----

func (pg UserSessions) get(ctx context.Context, by string, value any) (*entity.UserSession, error) {
	var buf strings.Builder
	buf.Grow(128 + len(database.TableUserSessions) + len(by))
	buf.WriteString("SELECT id, jwt_id, user_id, token_hash, ip_address, os_name, browser_name, device_type, expires_at, created_at FROM " + database.TableUserSessions + " WHERE ")
	buf.WriteString(by)
	buf.WriteString(" = $1")
	var session entity.UserSession
	err := pg.Pool.QueryRow(ctx, buf.String(), value).
		Scan(
			&session.ID,
			&session.JwtID,
			&session.UserID,
			&session.TokenHash,
			&session.IPAddress,
			&session.OSName,
			&session.BrowserName,
			&session.DeviceType,
			&session.ExpiresAt,
			&session.CreatedAt,
		)
	if err != nil {
		return nil, handlePgxError(fmt.Errorf("user sessions get by %s, %w", by, err))
	}
	return &session, nil
}

func (pg UserSessions) GetByID(ctx context.Context, id uuid.UUID) (*entity.UserSession, error) {
	return pg.get(ctx, "id", id)
}

func (pg UserSessions) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error) {
	return pg.get(ctx, "token_hash", tokenHash)
}

// ? -----+-----+----- Update -----+-----+-----

func (pg UserSessions) update(ctx context.Context, id uuid.UUID, col string, val any) error {
	var buf strings.Builder
	buf.Grow(31 + len(database.TableUserSessions) + len(col))
	buf.WriteString("UPDATE " + database.TableUserSessions + " SET ")
	buf.WriteString(col)
	buf.WriteString(" = $1 WHERE id = $2")

	cmdTag, err := pg.Pool.Exec(ctx, buf.String(), val, id)
	if err != nil {
		return handlePgxError(fmt.Errorf("user sessions update col %s, %w", col, err))
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}

func (pg UserSessions) UpdateTokenHashAndExpiry(ctx context.Context, id uuid.UUID, tokenHash string, expiry time.Time) error {
	cmdTag, err := pg.Pool.Exec(
		ctx,
		"UPDATE "+database.TableUserSessions+" SET token_hash = $1, expires_at = $2 WHERE id = $3",
		tokenHash, expiry, id,
	)
	if err != nil {
		return handlePgxError(fmt.Errorf("user sessions update token hash and expiry, %w", err))
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}

// ? -----+-----+----- Delete -----+-----+-----

func (pg UserSessions) delete(ctx context.Context, col string, val any) error {
	var buf strings.Builder
	buf.Grow(24 + len(database.TableUserSessions) + len(col))
	buf.WriteString("DELETE FROM " + database.TableUserSessions + " WHERE ")
	buf.WriteString(col)
	buf.WriteString(" = $1")
	cmdTag, err := pg.Pool.Exec(ctx, buf.String(), val)
	if err != nil {
		return handlePgxError(fmt.Errorf("user sessions delete by %s, %w", col, err))
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}

func (pg UserSessions) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return pg.delete(ctx, "id", id)
}

func (pg UserSessions) DeleteByJwtID(ctx context.Context, jwtID uuid.UUID) error {
	return pg.delete(ctx, "jwt_id", jwtID)
}
