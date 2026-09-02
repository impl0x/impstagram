package auth

import (
	"backend/internal/pkg/postgresql"
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ? INFO:
// implementation file for the repository
// ! Ownership and usage:
// owned by itself and implements the [Repository] interface with the [PostgresRepository] struct
// used by service indirectly behind repository

type PostgresRepository struct {
	Db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db}
}

// wraps any returned T, error to pass non nil error into handlePgxError
func wrapErrHandler[T any](t T, err error) (T, error) {
	if err != nil {
		return t, handlePgxError(err)
	}
	return t, nil
}

// ? ----+-----+-----Users table-----+-----+-----

func (pg PostgresRepository) findUser(ctx context.Context, one string, two any) (*userModel, error) {
	return wrapErrHandler(
		postgresql.Find[userModel](
			ctx,
			pg.Db,
			postgresql.QuerySelectAllWhere(
				tableUsers,
				one,
				two,
			),
		),
	)
}

func (pg PostgresRepository) findUserByID(ctx context.Context, userID uuid.UUID) (*userModel, error) {
	return pg.findUser(ctx, "id", userID)
}
func (pg PostgresRepository) findUserByChannel(ctx context.Context, channel authChannel, target string) (*userModel, error) {
	switch channel {
	case channelUsername:
		return wrapErrHandler(
			postgresql.Find[userModel](
				ctx,
				pg.Db,
				postgresql.QuerySelectAllInnerJoin(
					tableUsers,
					tableProfiles,
					"id",
					"user_id",
					"username",
					target,
				),
			),
		)
	case channelEmail, channelPhone:
		return pg.findUser(ctx, string(channel), target)
	default:
		return nil, fmt.Errorf("auth.PostgresRepository.findUserByChannel: %w in repository function call")
	}
}

// user parameter must only have populated values according to the db model constructor defined in models
//
// extra populated fields will not be inserted
func (pg PostgresRepository) createUser(ctx context.Context, user *userModel) (uuid.UUID, error) {
	return wrapErrHandler(
		postgresql.CreateAndReturnID(
			ctx,
			pg.Db,
			postgresql.QueryInsert(
				tableUsers,
				[]string{
					"email",
					"phone",
					"password_hash",
					"dob",
				},
				[]any{
					user.Email,
					user.Phone,
					user.PasswordHash,
					user.Dob,
				},
			),
		),
	)
}

func (pg PostgresRepository) updateUser(ctx context.Context, id uuid.UUID, one string, two any) error {
	return handlePgxError(
		postgresql.Update(
			ctx,
			pg.Db,
			postgresql.QueryUpdateOneWhere(
				tableUsers,
				"id",
				id,
				one,
				two,
			),
		),
	)
}
func (pg PostgresRepository) updateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return pg.updateUser(ctx, userID, "password_hash", passwordHash)
}
func (pg PostgresRepository) updateUserStatus(ctx context.Context, userID uuid.UUID, status accountStatus) error {
	return pg.updateUser(ctx, userID, "status", status)
}
func (pg PostgresRepository) updateUser2FA(ctx context.Context, userID uuid.UUID, twoFAs twoFAs) error {
	return pg.updateUser(ctx, userID, "two_fas", twoFAs)
}

func (pg PostgresRepository) enableTotp(ctx context.Context, userID uuid.UUID, secretKey string) error {
	return handlePgxError(
		postgresql.Update(
			ctx,
			pg.Db,
			postgresql.NewQuery(
				`UPDATE users SET totp_secret_key = $1 two_fas = array_append(two_fas, $2) WHERE id = $3`,
				secretKey,
				channelTOTP,
				userID,
			),
		),
	)
}

// ? ----+-----+-----User sessions table-----+-----+-----

func (pg PostgresRepository) findSession(ctx context.Context, one string, two any) (*userSessionModel, error) {
	session, err := postgresql.Find[userSessionModel](ctx, pg.Db, postgresql.QuerySelectAllWhere(tableUserSessions, one, two))
	if err != nil {
		return nil, handlePgxError(err)
	}
	return session, nil
}
func (pg PostgresRepository) findSessionByToken(ctx context.Context, tokenHash string) (*userSessionModel, error) {
	return pg.findSession(ctx, "token_hash", tokenHash)
}

func (pg PostgresRepository) createSession(ctx context.Context, session *userSessionModel) error {
	return handlePgxError(
		postgresql.Create(
			ctx,
			pg.Db,
			postgresql.QueryInsert(
				tableUserSessions,
				[]string{
					"jwt_id",
					"token_hash",
					"user_id",

					"ip_address",
					"os_name",
					"browser_name",
					"device_type",

					"expires_at",
				},
				[]any{
					session.JwtID,
					session.TokenHash,
					session.UserID,

					session.IPAddress,
					session.OSName,
					session.BrowserName,
					session.DeviceType,

					session.ExpiresAt,
				},
			),
		),
	)
}
// TODO: below all todo
func (pg PostgresRepository) updateSession(ctx context.Context, id uuid.UUID, one string, two any) error {
	query := `UPDATE user_sessions SET $1 = $2 WHERE id = $3`
	cmdTag, err := pg.Db.Exec(ctx, query, one, two, id)
	if err != nil {
		return handlePgxError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errRepoNoResults
	}
	return nil
}

func (pg PostgresRepository) updateSessionToken(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return pg.updateSession(ctx, id, "token_hash", tokenHash)
}
func (pg PostgresRepository) deleteSession(ctx context.Context, one string, two any) error {
	query := `DELETE FROM user_sessions WHERE $1 = $2`
	cmdTag, err := pg.Db.Exec(ctx, query, one, two)
	if err != nil {
		return handlePgxError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errRepoNoResults
	}
	return nil
}
func (pg PostgresRepository) deleteSessionByID(ctx context.Context, id uuid.UUID) error {
	return pg.deleteSession(ctx, "id", id)
}
func (pg PostgresRepository) deleteSessionByJwtID(ctx context.Context, jwtID uuid.UUID) error {
	return pg.deleteSession(ctx, "id", jwtID)
}

func handlePgxError(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return errRepoNoResults
	case errors.Is(err, pgx.ErrTooManyRows):
		return errRepoTooManyResults
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return err
	case errors.As(err, &pgErr):
		return fmt.Errorf("postgres: sql error, %w, Code: %s", err, pgErr.Code) // we let the error bubble to the handler where it will eventually be turned into a internal error
	default:
		return fmt.Errorf("postgres: unknown error, %w", err)
	}
}
