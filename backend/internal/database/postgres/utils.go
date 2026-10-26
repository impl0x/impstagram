package postgres

import (
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// calls the table specific after connect functions
func afterConnect(ctx context.Context, conn *pgx.Conn) error {
	// not returning the func call directly because in future we might
	// add more after connect functions, so letting the pattern stay
	err := afterConnectUsers(ctx, conn)
	if err != nil {
		return err
	}
	return nil
}

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
		switch pgErr.Code {
		case "23505":
			return repository.ErrAlreadyExists
		}
		return fmt.Errorf("postgres: sql error, %w, Code: %s", err, pgErr.Code) // we let the error bubble to the handler where it will eventually be turned into a internal error and be logged
	default:
		return fmt.Errorf("postgres: unknown error, %w", err)
	}
}

func colWriter(b *strings.Builder, col []string) {
	for i, c := range col {
		b.WriteString(c)
		if i != len(col)-1 {
			b.WriteByte(',')
		}
	}
}
func valWriter(b *strings.Builder, l int) {
	for i := range l {
		b.WriteByte('$')
		writeUint64(b, uint64(i+1))
		if i != l-1 {
			b.WriteByte(',')
		}
	}
}

func writeUint64(b *strings.Builder, n uint64) {
	var buf [20]byte // max length of max uint64
	pos := len(buf)

	if n == 0 {
		b.WriteByte('0')
		return
	}

	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}

	b.Write(buf[pos:])
}
