package app

import (
	"backend/internal/features/auth"
)

type App struct {
	Auth auth.Handler
}

func NewApp(auth auth.Handler) *App {
	return &App{auth}
}

func NewAuth() auth.Handler {
	return NewAuth()
}
