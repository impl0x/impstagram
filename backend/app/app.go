package app

import (
	"backend/internal/features/auth"
	"backend/internal/pkg/email"
	"net/http"
)

type App struct {
	Auth auth.Handler
}

func NewApp(auth auth.Handler) *App {
	return &App{auth}
}

func NewAuth() auth.Handler {
	repo:=auth.NewDemoRepository()
	service:=auth.NewService(repo, email.NewClient(http.DefaultClient))
	handler:=auth.NewHandler(service)
	return handler
}

