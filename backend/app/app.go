package app

import (
	"backend/internal/features/auth"
	"backend/internal/pkg/email"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Auth auth.Handler
}

func NewApp(auth auth.Handler) *App {
	return &App{auth}
}

func NewAuth(db *pgxpool.Pool) auth.Handler {
	return auth.NewHandler(auth.NewService(auth.NewPostgresRepository(db), email.NewClient(http.DefaultClient)))
}
