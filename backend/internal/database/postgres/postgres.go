package postgres

import (
	"backend/internal/config"
	"backend/pkg/postgres"

	"github.com/impl0x/go-utils/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Users        Users
	UserSessions UserSessions
	Profiles     Profiles
}

func New(c config.Postgres, log logger.Logger) (Postgres, error) {
	pg, err := postgres.New(
		c.URL,
		log,
		postgres.WithMaxPoolSize(c.PoolMax),
		postgres.WithPgxpoolConfig(func(c *pgxpool.Config) { c.AfterConnect = afterConnect }),
	)
	if err != nil {
		return Postgres{}, err
	}

	return Postgres{
		Users{pg},
		UserSessions{pg},
		Profiles{pg},
	}, nil
}
