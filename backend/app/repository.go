package app

import "backend/internal/repository"

// repositories used by the services
type repositories struct {
	user        repository.UserRepository
	userSession repository.UserSessionRepository
	profile     repository.ProfileRepository
}
