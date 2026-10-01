package database

import (
	"backend/archive/postgresql"
	"backend/internal/entity"
	"backend/pkg/postgres"
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

// ? INFO:
// implementation file for the repository
// ! Ownership and usage:
// owned by itself and implements the [Repository] interface with the [UserPostgresRepository] struct
// used by service indirectly behind repository

type UserPostgresRepository struct {
	postgres.Postgres
}

func NewUserPostgresRepository(pg postgres.Postgres) UserPostgresRepository {
	return UserPostgresRepository{pg}
}

// ? ----+-----+-----Users table-----+-----+-----
//
// user parameter must only have populated values according to the db model constructor defined in models
//
// extra populated fields will not be inserted
func (pg UserPostgresRepository) CreateUser(ctx context.Context, user *entity.User, username string) (uuid.UUID, error) {
	if user == nil {
		panic("user is nil")
	}
	tx, err := pg.Pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, handlePgxError(err)
	}
	defer tx.Rollback(ctx)
	cols := newArgBuilder[string]()
	vals := newArgBuilder[any]()
	// if provided a ID then don't rely on the database to make one again, use the id given
	if user.ID != uuid.Nil() {
		cols.add("id")
		vals.add(user.ID)
	}
	// we dont actually care if the user sent a created_at or a updated_at because those are hardcoded here to be the current timestamp by the database.
	cols.add("email", "phone", "password_hash", "dob", "status", "totp_secret_key", "two_fas")
	vals.add(user.Email, user.Phone, user.PasswordHash, user.Dob, user.TotpSecretKey, user.TwoFAs)
	sql, args, err := pg.Builder.
		Insert(tableUsers).
		Columns(argBuilderJoinArgs(cols.unwrap())).
		Values(vals.unwrap()...).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil(), fmt.Errorf("UserRepo - CreateUser - pg.Builder: %w", err)
	}

	var userID uuid.UUID
	err = tx.QueryRow(ctx, sql, args...).Scan(&userID)
	if err != nil {
		return uuid.UUID{}, handlePgxError(err)
	}
	sql, args, err = pg.Builder.
		Insert(tableProfiles).
		Columns("user_id").
		Values(userID, username).
		ToSql()
	_, err = tx.Exec(ctx, sql, args)
	if err != nil {
		return uuid.UUID{}, handlePgxError(err)
	}
	return userID, handlePgxError(tx.Commit(ctx))
}

func (pg UserPostgresRepository) getUser(ctx context.Context, by string, value any) (*entity.User, error) {
	sql, args, err := pg.Builder.
		Select("id, email, phone, password_hash, dob, status, totp_secret_key, two_fas, created_at, updated_at").
		From(tableUsers).
		Where(squirrel.Eq{by: value}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserRepo - getUser - pg.Builder: %w", err)
	}
	var user entity.User
	err = pg.Pool.QueryRow(ctx, sql, args...).
		Scan(
			&user.ID,
			&user.Email,
			&user.Phone,
			&user.PasswordHash,
			&user.Dob,
			&user.Status,
			&user.TotpSecretKey,
			&user.TwoFAs,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
	if err != nil {
		return nil, handlePgxError(err)
	}
	return &user, nil
}

func (pg UserPostgresRepository) GetUserByID(ctx context.Context, ID uuid.UUID) (*entity.User, error) {
	return pg.getUser(ctx, "id", ID)
}
func (pg UserPostgresRepository) GetUserByAuthChannel(ctx context.Context, channel entity.AuthChannel, target string) (*entity.User, error) {
	switch channel {
	case entity.ChannelUsername:
		sql, args, err := pg.Builder.
			Select("t1.id, t1.email, t1.phone, t1.password_hash, t1.dob, t1.status, t1.totp_secret_key, t1.two_fas, t1.created_at, t1.updated_at").
			From(tableUsers + " t1").
			Join(tableProfiles + " t2 ON t1.id = t2.user_id").
			Where(squirrel.Eq{"t2.username": target}).
			ToSql()
		if err != nil {
			return nil, fmt.Errorf("UserRepo - getUser by username - pg.Builder: %w", err)
		}
		var user entity.User
		err = pg.Pool.QueryRow(ctx, sql, args...).
			Scan(
				&user.ID,
				&user.Email,
				&user.Phone,
				&user.PasswordHash,
				&user.Dob,
				&user.Status,
				&user.TotpSecretKey,
				&user.TwoFAs,
				&user.CreatedAt,
				&user.UpdatedAt,
			)
		if err != nil {
			return nil, handlePgxError(err)
		}
		return &user, nil
	case entity.ChannelEmail, entity.ChannelPhone:
		return pg.getUser(ctx, string(channel), target)
	default:
		panic("invalid channel, provided channel: " + string(channel))

	}
}

func (pg UserPostgresRepository) UpdateUser(ctx context.Context, id uuid.UUID, col string, val any) error {
	sql, args, err := pg.Builder.
		Update(tableUsers).
		Set(col, val).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - UpdateUser - pg.Builder: %w", err)
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

// both user account and profile is deleted, permanently.
func (pg UserPostgresRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	sql, args, err := pg.Builder.
		Delete(tableUsers).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - DeleteUser - pg.Builder: %w", err)
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

// ? ----+-----+-----User sessions table-----+-----+-----
func (pg UserPostgresRepository) CreateSession(ctx context.Context, session *entity.UserSession) error {
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
		Into(tableUserSessions).
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

func (pg UserPostgresRepository) getSession(ctx context.Context, by string, value any) (*entity.UserSession, error) {
	sql, args, err := pg.Builder.
		Select("id, jwt_id, user_id, token_hash, ip_address, os_name, browser_name, device_type, expires_at, created_at").
		From(tableUserSessions).
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

func (pg UserPostgresRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error) {
	return pg.getSession(ctx, "token_hash", tokenHash)
}

func (pg UserPostgresRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*entity.UserSession, error) {
	return pg.getSession(ctx, "id", id)
}

func (pg UserPostgresRepository) UpdateSession(ctx context.Context, id uuid.UUID, col string, val any) error {
	sql, args, err := pg.Builder.
		Update(tableUserSessions).
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
// TODO: rest are to be done
func (pg UserPostgresRepository) updateSessionToken(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return pg.updateSession(ctx, id, "token_hash", tokenHash)
}
func (pg UserPostgresRepository) deleteSession(ctx context.Context, one string, two any) error {
	return handlePgxError(postgresql.Delete(ctx, pg.Db, postgresql.QueryDeleteWhere(tableUserSessions, one, two)))
}
func (pg UserPostgresRepository) deleteSessionByID(ctx context.Context, id uuid.UUID) error {
	return pg.deleteSession(ctx, "id", id)
}
func (pg UserPostgresRepository) deleteSessionByJwtID(ctx context.Context, jwtID uuid.UUID) error {
	return pg.deleteSession(ctx, "id", jwtID)
}
