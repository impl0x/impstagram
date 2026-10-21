package postgres

import (
	"backend/internal/database"
	"backend/internal/entity"
	"backend/pkg/postgres"
	"context"
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

func NewUserSessions(pg postgres.Postgres) UserSessions {
	return UserSessions{pg}
}

// ? -----+-----+----- Create -----+-----+-----

func (pg UserSessions) Create(ctx context.Context, session *entity.UserSession) error {
	totalLen := 85
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
	// ignore created_at even if provided. hardcoded to not set it via service call
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
		handlePgxError(err)
	}
	return nil
}

// ? -----+-----+----- Get -----+-----+-----

func (pg UserSessions) get(ctx context.Context, by string, value any) (*entity.UserSession, error) {
	var session entity.UserSession
	err := pg.Pool.QueryRow(
		ctx,
		"SELECT id, jwt_id, user_id, token_hash, ip_address, os_name, browser_name, device_type, expires_at, created_at FROM "+
			database.TableUserSessions+
			" WHERE $1 = $2", by, value,
	).
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
		return nil, handlePgxError(err)
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
	cmdTag, err := pg.Pool.Exec(ctx, "UPDATE "+database.TableUserSessions+" SET $1 = $2 WHERE id = $3", col, val, id)
	if err != nil {
		return handlePgxError(err)
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
		return handlePgxError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}

// ? -----+-----+----- Delete -----+-----+-----

func (pg UserSessions) delete(ctx context.Context, col string, val any) error {
	cmdTag, err := pg.Pool.Exec(ctx, "DELETE FROM "+database.TableUserSessions+" WHERE $1 = $2", col, val)
	if err != nil {
		return handlePgxError(err)
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
