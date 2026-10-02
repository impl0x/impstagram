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

// INFO:
// contains users table methods

type Users struct {
	postgres.Postgres
}

func NewUsers(pg postgres.Postgres) Users {
	return Users{pg}
}

// ? ----+-----+-----Users table-----+-----+-----
//
// user parameter must only have populated values according to the db model constructor defined in models
//
// extra populated fields will not be inserted
func (pg Users) CreateUser(ctx context.Context, user *entity.User, username string) (uuid.UUID, error) {
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
		Insert(database.TableUsers).
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
		Insert(database.TableProfiles).
		Columns("user_id").
		Values(userID, username).
		ToSql()
	_, err = tx.Exec(ctx, sql, args)
	if err != nil {
		return uuid.UUID{}, handlePgxError(err)
	}
	return userID, handlePgxError(tx.Commit(ctx))
}

func (pg Users) getUser(ctx context.Context, by string, value any) (*entity.User, error) {
	sql, args, err := pg.Builder.
		Select("id, email, phone, password_hash, dob, status, totp_secret_key, two_fas, created_at, updated_at").
		From(database.TableUsers).
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

func (pg Users) GetUserByID(ctx context.Context, ID uuid.UUID) (*entity.User, error) {
	return pg.getUser(ctx, "id", ID)
}
func (pg Users) GetUserByAuthChannel(ctx context.Context, channel entity.AuthChannel, target string) (*entity.User, error) {
	switch channel {
	case entity.ChannelUsername:
		sql, args, err := pg.Builder.
			Select("t1.id, t1.email, t1.phone, t1.password_hash, t1.dob, t1.status, t1.totp_secret_key, t1.two_fas, t1.created_at, t1.updated_at").
			From(database.TableUsers + " t1").
			Join(database.TableProfiles + " t2 ON t1.id = t2.user_id").
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

func (pg Users) UpdateUser(ctx context.Context, id uuid.UUID, col string, val any) error {
	sql, args, err := pg.Builder.
		Update(database.TableUsers).
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
func (pg Users) DeleteUser(ctx context.Context, id uuid.UUID) error {
	sql, args, err := pg.Builder.
		Delete(database.TableUsers).
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
