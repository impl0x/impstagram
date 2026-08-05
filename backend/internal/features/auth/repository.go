package auth

import "github.com/google/uuid"

type Repository interface {
	FindByEmail(email string) (*User, error)

	Create(user *User) error

	Get(id uuid.UUID) (*User, error)
}