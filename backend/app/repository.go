package app

import (
	"backend/internal/database/postgres"
	"backend/internal/repository"
)

// repositories used by the services
type repositories struct {
	user        repository.UserRepository
	userSession repository.UserSessionRepository
	profile     repository.ProfileRepository
}

// returns repositories made using postgres implementations
func newRepositoriesFromPostgres(pg postgres.Postgres) repositories {
	return repositories{
		user:        pg.Users,
		userSession: pg.UserSessions,
		profile:     pg.Profiles,
	}
}
