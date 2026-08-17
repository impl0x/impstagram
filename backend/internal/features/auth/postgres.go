package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	Db *pgxpool.Pool
}

func NewPostgresRepository() PostgresRepository {
	return PostgresRepository{}
}

func (pg PostgresRepository) findUser(ctx context.Context, one string, two any) (*user, error){
	rows, err:=pg.Db.Query(ctx, `SELECT * FROM users WHERE $1 = $2`,one,two)
	if err!=nil{
		return nil, pg.handleError(err)
	}
	user, err:=pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[user])
	if err!=nil{
		return nil, pg.handleError(err)
	}
	return &user, nil
}

func (pg PostgresRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*user, error) {
	return pg.findUser(ctx, "id", userID)
}
func (pg PostgresRepository) FindUserByChannel(ctx context.Context, channel authChannel, target string) (*user, error) {
	return pg.findUser(ctx, string(channel), target)
}
// todo
func (pg PostgresRepository) CreateUser(ctx context.Context, user *user) error {
	return nil
}
func (pg PostgresRepository) UpdateUser(ctx context.Context, userID uuid.UUID, updatedUser *user) error {
	return nil
}

// user_sessions table{}
func (pg PostgresRepository) FindSessionByToken(ctx context.Context, tokenHash string) (*userSession, error) {
	return nil, nil
}
func (pg PostgresRepository) CreateSession(ctx context.Context, session *userSession) error {
	return nil
}
func (pg PostgresRepository) UpdateSessionToken(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return nil
}
func (pg PostgresRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (pg PostgresRepository) DeleteSessionByJwtID(ctx context.Context, jwtID uuid.UUID) error {
	return nil
}

var (
	errNoResults      = errors.New("no results found")
	errTooManyResults = errors.New("too many results")
)

func (pg PostgresRepository) handleError(err error) error {
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
	case errors.Is(err, pgErr):
		return fmt.Errorf("repository: sql error, %w", err) // we let the error bubble to the handler where it will eventually be turned into a internal error
	default:
		return fmt.Errorf("repository: unknown error, %w", err) 
	}
}
