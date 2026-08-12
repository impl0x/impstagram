package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	Db *pgxpool.Pool
}

var demoUser = &User{
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
func (pg PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return demoUser, nil
}
func (pg PostgresRepository) FindByChannel(ctx context.Context, channel authChannel, target string) (*User, error) {
	return demoUser, nil
}
func (pg PostgresRepository) Create(ctx context.Context, user *User) error {
	return nil
}

