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
		if pgErr.Code == "23505" { // unique violation
			return repository.ErrAlreadyExists
		}
		return fmt.Errorf("postgres: sql error, %w, Code: %s", err, pgErr.Code) // we let the error bubble to the handler where it will eventually be turned into a internal error and be logged
	default:
		return fmt.Errorf("postgres: unknown error, %w", err)
	}
}

const _argBuilderDefaultCap uint8 = 16

type argBuilder struct {
	cols []string
	vals []any
}

func newArgBuilder() argBuilder {
	return argBuilder{
		make([]string, 0, _argBuilderDefaultCap),
		make([]any, 0, _argBuilderDefaultCap),
	}
}

// grows the underlying slice to the capacity provided
func (ag argBuilder) withCap(cap int) argBuilder {
	return argBuilder{slices.Grow(ag.cols, cap), slices.Grow(ag.vals, cap)}
}

func (ag argBuilder) add(col string, val any) {
	ag.cols = append(ag.cols, col)
	ag.vals = append(ag.vals, val)
}

func (ag argBuilder) addCols(col ...string) {
	ag.cols = append(ag.cols, col...)
}
func (ag argBuilder) addVals(val ...any) {
	ag.vals = append(ag.vals, val...)
}

func (ag argBuilder) unwrapCols() string {
	return strings.Join(ag.cols, ", ")
}

func (ag argBuilder) getVals() []any {
	return ag.vals
}
