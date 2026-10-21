package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/impl0x/go-utils/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
	log     logger.Logger
}

func New(url string, log logger.Logger, opts ...Option) (Postgres, error) {
	options := defaultOptions
	var err error
	options.pgxpoolConfig, err = pgxpool.ParseConfig(url)
	if err != nil {
		return Postgres{}, fmt.Errorf("postgres - New - pgxpool.ParseConfig: %w", err)
	}

	// apply custom options
	for _, op := range opts {
		op(&options)
	}

	var pool *pgxpool.Pool
	// attempt to connect to database
	for options.connAttempts > 0 {
		pool, err = pgxpool.NewWithConfig(context.Background(), options.pgxpoolConfig)
		if err == nil {
			break
		}
		log.Info("Postgres is trying to connect, attempts left: " + strconv.FormatUint(uint64(options.connAttempts), 10))

		time.Sleep(options.connTimeout)

		options.connAttempts--
	}
	// if fail to connect after all attempts
	if err != nil {
		return Postgres{}, fmt.Errorf("postgres - New - connAttempts == 0: %w", err)
	}

	return Postgres{
		squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		pool,
		log,
	}, nil
}

func (pg Postgres) Close() {
	if pg.Pool != nil {
		pg.Pool.Close()
	}
}
