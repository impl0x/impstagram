package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	Db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) postgresRepository {
	return postgresRepository{db}
}

// users table
func (pg postgresRepository) findUser(ctx context.Context, one string, two any) (*user, error) {
	rows, err := pg.Db.Query(ctx, `SELECT * FROM users WHERE $1 = $2`, one, two)
	if err != nil {
		return nil, pg.handleError(err)
	}
	defer rows.Close()
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[user])
	if err != nil {
		return nil, pg.handleError(err)
	}
	return &user, nil
}

func (pg postgresRepository) findUserByID(ctx context.Context, userID uuid.UUID) (*user, error) {
	return pg.findUser(ctx, "id", userID)
}
func (pg postgresRepository) findUserByChannel(ctx context.Context, channel authChannel, target string) (*user, error) {
	return pg.findUser(ctx, string(channel), target)
}

func (pg postgresRepository) createUser(ctx context.Context, user *user) (uuid.UUID, error) {
	query := `INSERT INTO users (username, email, phone, totp_secret_key, password_hash, dob, status, two_fas, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	var pgID pgtype.UUID
	err := pg.Db.QueryRow(ctx, query, user.Username, user.Email, user.Phone, user.TotpSecretKey, user.PasswordHash, user.Dob, user.Status, user.TwoFAs, user.CreatedAt, user.UpdatedAt).Scan(&pgID)
	if err != nil {
		return uuid.UUID{}, pg.handleError(err)
	}
	return uuid.UUID(pgID.Bytes), nil
}

func (pg postgresRepository) updateUser(ctx context.Context, id uuid.UUID, one string, two any) error {
	query := `UPDATE users SET $1 = $2 WHERE id = $3`
	cmdTag, err := pg.Db.Exec(ctx, query, one, two, id)
	if err != nil {
		return pg.handleError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errNoResults
	}
	return nil
}
func (pg postgresRepository) updateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return pg.updateUser(ctx, userID, "password_hash", passwordHash)
}
func (pg postgresRepository) updateUserStatus(ctx context.Context, userID uuid.UUID, status accountStatus) error {
	return pg.updateUser(ctx, userID, "status", status)
}
func (pg postgresRepository) updateUser2FA(ctx context.Context, userID uuid.UUID, twoFAs twoFAs) error {
	return pg.updateUser(ctx, userID, "two_fas", twoFAs)
}

// user_sessions table{}

func (pg postgresRepository) findSession(ctx context.Context, one string, two any) (*userSession, error) {
	rows, err := pg.Db.Query(ctx, `SELECT * FROM user_sessions WHERE $1 = $2`, one, two)
	if err != nil {
		return nil, pg.handleError(err)
	}
	defer rows.Close()
	userSesh, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userSession])
	if err != nil {
		return nil, pg.handleError(err)
	}
	return &userSesh, nil
}

func (pg postgresRepository) findSessionByToken(ctx context.Context, tokenHash string) (*userSession, error) {
	return pg.findSession(ctx, "token_hash", tokenHash)
}

func (pg postgresRepository) createSession(ctx context.Context, session *userSession) error {
	query := `INSERT INTO user_sessions (jwt_id, token_hash, user_id, ip_address, os_name, browser_name, device_type, expires_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := pg.Db.Exec(ctx, query, session.JwtID, session.TokenHash, session.UserID, session.IPAddress, session.OSName, session.BrowserName, session.DeviceType, session.ExpiresAt, session.DeviceType)
	if err != nil {
		return pg.handleError(err)
	}
	return nil
}

func (pg postgresRepository) updateSession(ctx context.Context, id uuid.UUID, one string, two any) error {
	query := `UPDATE user_sessions SET $1 = $2 WHERE id = $3`
	cmdTag, err := pg.Db.Exec(ctx, query, one, two, id)
	if err != nil {
		return pg.handleError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errNoResults
	}
	return nil
}

func (pg postgresRepository) updateSessionToken(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return pg.updateSession(ctx, id, "token_hash", tokenHash)
}
func (pg postgresRepository) deleteSession(ctx context.Context, one string, two any) error {
	query := `DELETE FROM user_sessions WHERE $1 = $2`
	cmdTag, err := pg.Db.Exec(ctx, query, one, two)
	if err != nil {
		return pg.handleError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errNoResults
	}
	return nil
}
func (pg postgresRepository) deleteSessionByID(ctx context.Context, id uuid.UUID) error {
	return pg.deleteSession(ctx, "id", id)
}
func (pg postgresRepository) deleteSessionByJwtID(ctx context.Context, jwtID uuid.UUID) error {
	return pg.deleteSession(ctx, "id", jwtID)
}

func (pg postgresRepository) handleError(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return errNoResults
	case errors.Is(err, pgx.ErrTooManyRows):
		return errTooManyResults
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded // unwrap, we need to only return unwrapped errors if the handler is not going to raise a internal server error with them
	case errors.Is(err, context.Canceled):
		return context.Canceled // unwrap
	case errors.As(err, &pgErr):
		return fmt.Errorf("repository: sql error, %w, Code: %s", err, pgErr.Code) // we let the error bubble to the handler where it will eventually be turned into a internal error
	default:
		return fmt.Errorf("repository: unknown error, %w", err)
	}
}
