package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	Db *pgxpool.Pool
}

func NewPostgresRepository(Db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{Db: Db}
}

var demoUser = &user{
	uuid.New(),
	"test",
	"email@email.com",
	"1234567890",
	"abcedxyz",
	"passwhash",
	"dob",
	statusVerified,
	nil,
}

// todos
func (pg PostgresRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*user, error) {
	return demoUser, nil
}
func (pg PostgresRepository) FindUserByChannel(ctx context.Context, channel authChannel, target string) (*user, error) {
	return demoUser, nil
}
func (pg PostgresRepository) CreateUser(ctx context.Context, user *user) error {
	return nil
}
func (pg PostgresRepository) UpdateUser(ctx context.Context, userID uuid.UUID, updatedUser *user) error {
	return nil
}

var demoUserSession = &userSession{
	uuid.New(),
	uuid.New(),
	"", uuid.New(), "", "", "", "", time.Now().Add(time.Hour), time.Now(),
}

// user_sessions table{}
func (pg PostgresRepository) FindSessionByToken(ctx context.Context, tokenHash string) (*userSession, error) {
	return demoUserSession, nil
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
