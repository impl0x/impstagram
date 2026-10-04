package postgres

import (
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// returns pgx errors mapped to repository errors and handles unknown errors as well
func handlePgxError(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return repository.ErrNoResults
	case errors.Is(err, pgx.ErrTooManyRows):
		return repository.ErrTooManyResults
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return err
	case errors.As(err, &pgErr):
		return fmt.Errorf("postgres: sql error, %w, Code: %s", err, pgErr.Code) // we let the error bubble to the handler where it will eventually be turned into a internal error and be logged
	default:
		return fmt.Errorf("postgres: unknown error, %w", err)
	}
}

const _argBuilderDefaultCap uint8 = 16

// utility struct to build sql values/cols lines.
type argBuilder[T any] struct {
	slice []T
}

// returns a new arg builder with the default capacity [_argBuilderDefaultCap] for the underlying slice
func newArgBuilder[T any]() argBuilder[T] {
	return argBuilder[T]{make([]T, 0, _argBuilderDefaultCap)}
}

// grows the underlying slice to the capacity provided
func (ag argBuilder[T]) withCap(cap uint8) argBuilder[T] {
	return argBuilder[T]{slices.Grow(ag.slice, int(cap))}
}

func (ag argBuilder[T]) add(elem ...T) {
	ag.slice = append(ag.slice, elem...)
}

func (ag argBuilder[T]) unwrap() []T {
	return ag.slice
}

func argBuilderJoinArgs(elems []string) string {
	return strings.Join(elems, ", ")
}
