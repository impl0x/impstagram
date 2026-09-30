package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type options struct {
	maxPoolSize   uint32
	connAttempts  uint32
	connTimeout   time.Duration
	pgxpoolConfig *pgxpool.Config
}

var defaultOptions = options{
	maxPoolSize:   1,
	connAttempts:  5,
	connTimeout:   time.Second * 2,
	pgxpoolConfig: nil, // is set after parsing url
}

type Option func(*options) // a function which modifies an options struct

// Maximum of connections the pool is allowed to hold at a time
func MaxPoolSize(size uint32) Option {
	return func(o *options) {
		o.maxPoolSize = size
	}
}

// Number of attempts before the driver gives up and returns an error
func ConnAttempts(attempts uint32) Option {
	return func(o *options) {
		o.connAttempts = attempts
	}
}

// Sets the connection timeout value before the next attempt
func ConnTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.connTimeout = timeout
	}
}

// The given function "fn" is provided a [*pgxpool.Config] which can then be written to and configured as per wish.
// Argument [*pgxpool.Config] is always non-nil
func PgxpoolConfig(fn func(*pgxpool.Config)) Option {
	return func(o *options) {
		fn(o.pgxpoolConfig)
	}
}
