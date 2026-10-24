package postgres

import (
	"backend/internal/database"
	"backend/internal/entity"
	"backend/pkg/postgres"
	"context"
	"fmt"
	"strings"
	"uuid"

	"github.com/jackc/pgx/v5"
)

// INFO:
// contains users table methods

type Users struct {
	postgres.Postgres
}

// registers the enums needed for users table
func afterConnectUsers(ctx context.Context, conn *pgx.Conn) error {
	// enum
	enumType, err := conn.LoadType(ctx, "auth_channel")
	if err != nil {
		return err
	}
	conn.TypeMap().RegisterType(enumType)

	// enum array
	arrayType, err := conn.LoadType(ctx, "_auth_channel")
	if err != nil {
		return err
	}
	conn.TypeMap().RegisterType(arrayType)

	return nil
}

// ? -----+-----+----- Create -----+-----+-----
//
// user parameter must only have populated values according to the db model constructor defined in models
//
// extra populated fields will not be inserted
func (pg Users) Create(ctx context.Context, user *entity.User, username string) (uuid.UUID, error) {
	if user == nil {
		panic("user is nil")
	}
	tx, err := pg.Pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, handlePgxError(fmt.Errorf("user create transaction start, %w", err))
	}
	defer tx.Rollback(ctx)

	totalLen := 113 // calculated number of bytes
	cols := make([]string, 0, 8)
	vals := make([]any, 0, 8)
	// if provided a ID then don't rely on the database to make one again, use the id given
	if user.ID != uuid.Nil() {
		pg.Log.Warn("DB: Providing ID in create user is not good, let the database create that by default for you")
		cols = append(cols, "id")
		vals = append(vals, user.ID)
		totalLen += 6
	}
	if user.Status != "" {
		pg.Log.Warn("DB: Providing Status in create user is not good, let the database default that by default for you")
		cols = append(cols, "status")
		vals = append(vals, user.Status)
		totalLen += 10
	}
	// we dont actually care if the user sent a created_at or a updated_at because those are hardcoded here to be the current timestamp by the database.
	cols = append(cols, "email", "phone", "password_hash", "dob", "totp_secret_key", "two_fas")
	vals = append(vals, user.Email, user.Phone, user.PasswordHash, user.Dob, user.TotpSecretKey, user.TwoFAs)

	var buf strings.Builder
	buf.Grow(totalLen) // should fit according to our calculations
	buf.WriteString("INSERT INTO " + database.TableUsers + " (")
	colWriter(&buf, cols)
	buf.WriteString(") VALUES (")
	valWriter(&buf, len(vals))
	buf.WriteString(") RETURNING id")

	var userID uuid.UUID
	err = tx.QueryRow(ctx, buf.String(), vals...).Scan(&userID)
	if err != nil {
		return uuid.UUID{}, handlePgxError(fmt.Errorf("user create, %w", err))
	}

	_, err = tx.Exec(ctx, "INSERT INTO "+database.TableProfiles+" (user_id,username) VALUES ($1,$2)", userID, username)

	if err != nil {
		return uuid.UUID{}, handlePgxError(fmt.Errorf("user profile create, %w", err))
	}
	err = tx.Commit(ctx)
	if err != nil {
		return uuid.UUID{}, handlePgxError(fmt.Errorf("user create transaction commit, %w", err))
	}
	return userID, nil
}

// ? -----+-----+----- Get -----+-----+-----

// wrapper for get methods
func (pg Users) get(ctx context.Context, by string, value any) (*entity.User, error) {
	var buf strings.Builder
	buf.Grow(130 + len(by))
	buf.WriteString("SELECT id, email, phone, password_hash, dob::text, status, totp_secret_key, two_fas, created_at, updated_at FROM " + database.TableUsers + " WHERE ")
	buf.WriteString(by)
	buf.WriteString(" = $1")
	var user entity.User
	err := pg.Pool.QueryRow(ctx, buf.String(), value).
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
		return nil, handlePgxError(fmt.Errorf("user get by %s, %w", by, err))
	}
	return &user, nil
}

func (pg Users) GetByID(ctx context.Context, ID uuid.UUID) (*entity.User, error) {
	return pg.get(ctx, "id", ID)
}
func (pg Users) GetByAuthChannel(ctx context.Context, channel entity.AuthChannel, target string) (*entity.User, error) {
	switch channel {
	case entity.ChannelUsername:
		var user entity.User
		err := pg.Pool.QueryRow(
			ctx,
			"SELECT t1.id, t1.email, t1.phone, t1.password_hash, t1.dob, t1.status, t1.totp_secret_key, t1.two_fas, t1.created_at, t1.updated_at FROM "+
				database.TableUsers+
				" t1 JOIN "+
				database.TableProfiles+
				" t2 ON t1.id = t2.user_id WHERE t2.username = $1",
			target,
		).
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
			return nil, handlePgxError(fmt.Errorf("user get by username, %w", err))
		}
		return &user, nil
	case entity.ChannelEmail, entity.ChannelPhone:
		return pg.get(ctx, string(channel), target)
	default:
		panic("postgres: invalid channel, provided channel: " + string(channel))
	}
}

// ? -----+-----+----- Update -----+-----+-----

// wrapper for update methods
func (pg Users) update(ctx context.Context, id uuid.UUID, col string, val any) error {
	var buf strings.Builder
	buf.Grow(36 + len(col))
	buf.WriteString("UPDATE " + database.TableUsers + " SET ")
	buf.WriteString(col)
	buf.WriteString(" = $1 WHERE id = $2")

	cmdTag, err := pg.Pool.Exec(ctx, buf.String(), val, id)
	if err != nil {
		return handlePgxError(fmt.Errorf("user update col %s, %w", col, err))
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}

func (pg Users) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus) error {
	return pg.update(ctx, id, "status", status)
}

func (pg Users) UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error {
	return pg.update(ctx, id, "password_hash", hash)
}
func (pg Users) UpdateTwoFAs(ctx context.Context, id uuid.UUID, twoFAs []entity.AuthChannel) error {
	return pg.update(ctx, id, "two_fas", twoFAs)
}
func (pg Users) UpdateAddTotp(ctx context.Context, id uuid.UUID, secretKey string) error {
	cmdTag, err := pg.Pool.Exec(ctx, "UPDATE "+database.TableUsers+" SET two_fas = array_append(two_fas, 'totp'), totp_secret_key = $1 WHERE id = $2", secretKey, id)
	if err != nil {
		return handlePgxError(fmt.Errorf("user update add totp, %w", err))
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}

// ? -----+-----+----- Delete -----+-----+-----

// both user account and profile is deleted, permanently.
func (pg Users) Delete(ctx context.Context, id uuid.UUID) error {
	cmdTag, err := pg.Pool.Exec(ctx, "DELETE FROM "+database.TableUsers+" WHERE id = $1", id)
	if err != nil {
		return handlePgxError(fmt.Errorf("user delete, %w", err))
	}
	if cmdTag.RowsAffected() == 0 {
		return handlePgxError(pgx.ErrNoRows)
	}
	return nil
}
