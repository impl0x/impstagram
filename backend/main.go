package main

import (
	"backend/internal/features/auth"
	"backend/internal/pkg/email"
	"backend/internal/pkg/jwt"
	"backend/internal/utils"

	"github.com/impl0x/mo"
)

func main() {
	m := mo.New()
	m.HTTPErrorHandler = utils.CustomErrorHandler

	authPostgres := auth.PostgresRepository{Db: nil}
	authService := auth.Service{Jwt: jwt.Jwt{},Email: email.NewClient("fakeapikey"),Repo: authPostgres}
	authHandler := auth.Handler{Service: &authService}

	v1Group:=m.Group("/api/v1")
	v1Group.POST("/login",authHandler.Login)
	
	m.Start(":8080")

}
