package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

// todos
func (pg *PostgresRepository) FindByUsername(ctx context.Context, username string) (User, error) {

}
func (pg *PostgresRepository) FindByEmail(email string) (User, error) {

}
func (pg *PostgresRepository) FindByPhone(phone string) (User, error) {

}
func (pg *PostgresRepository) Create(user User) error {

}
func (pg *PostgresRepository) Get(id uuid.UUID) (User, error) {

}
