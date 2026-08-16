package app

import (
	"backend/internal/features/auth"
	"backend/internal/pkg/email"
	"net/http"
)

func NewAuth() auth.Handler {
	return auth.NewHandler(auth.NewService(auth.NewPostgresRepository(nil), email.NewClient(http.DefaultClient)))
}
