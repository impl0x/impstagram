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
// Contains user_sessions table methods

type UserSessions struct {
	postgres.Postgres
}

func NewUserSessions(pg postgres.Postgres) UserSessions {
	return UserSessions{pg}
}

// ? ----+-----+-----User sessions table-----+-----+-----
func (pg UserSessions) Create(ctx context.Context, session *entity.UserSession) error {
	if session == nil {
		panic("session is nil")
	}
	cols := newArgBuilder[string]()
	vals := newArgBuilder[any]()
	if session.ID != uuid.Nil() {
		cols.add("id")
		vals.add(session.ID)
	}
	// ignore created_at even if provided. hardcoded to not set it via service call
	cols.add("jwt_id", "user_id", "token_hash", "os_name", "browser_name", "device_type", "expires_at")
	vals.add(session.JwtID, session.UserID, session.TokenHash, session.OSName, session.BrowserName, session.DeviceType, session.ExpiresAt)
	sql, args, err := pg.Builder.
		Insert(argBuilderJoinArgs(cols.unwrap())).
		Into(database.TableUserSessions).
		Values(vals.unwrap()...).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - createSession - pg.Builder: %w", err)
	}
	_, err = pg.Pool.Exec(ctx, sql, args...)
	if err != nil {
		handlePgxError(err)
	}
	return nil
}

func (pg UserSessions) get(ctx context.Context, by string, value any) (*entity.UserSession, error) {
	sql, args, err := pg.Builder.
		Select("id, jwt_id, user_id, token_hash, ip_address, os_name, browser_name, device_type, expires_at, created_at").
		From(database.TableUserSessions).
		Where(squirrel.Eq{by: value}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserRepo - getSession - pg.Builder: %w", err)
	}
	var session entity.UserSession
	err = pg.Pool.QueryRow(ctx, sql, args...).
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

func (pg UserSessions) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error) {
	return pg.get(ctx, "token_hash", tokenHash)
}

func (pg UserSessions) GetByID(ctx context.Context, id uuid.UUID) (*entity.UserSession, error) {
	return pg.get(ctx, "id", id)
}

func (pg UserSessions) Update(ctx context.Context, id uuid.UUID, col string, val any) error {
	sql, args, err := pg.Builder.
		Update(database.TableUserSessions).
		Set(col, val).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - UpdateSession - pg.Builder")
	}
	cmdTag, err := pg.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return handlePgxError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}

func (pg UserSessions) Delete(ctx context.Context, by string, value any) error {
	sql, args, err := pg.Builder.
		Delete(database.TableUserSessions).
		Where(squirrel.Eq{by: value}).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - UpdateSession - pg.Builder")
	}
	cmdTag, err := pg.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return handlePgxError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}
