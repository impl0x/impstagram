package main

import (
	"backend/app"
	"backend/internal/config"
	"backend/internal/utils"
	"context"
	"log"

	"github.com/impl0x/mo"
	"github.com/impl0x/mo/middlewares"
	"github.com/jackc/pgx/v5/pgxpool"
)
// to change later
func main() {
	m:=mo.New()
	m.HTTPErrorHandler=utils.CustomErrorHandler
	m.AddPostMiddleware(middlewares.LoggerWithResponseCode)
	db,err:=pgxpool.New(context.Background(), config.DbConnStr)
	if err!=nil{
		log.Fatal(err)
	}
	app:=app.NewApp(app.NewAuth(db))
	
	authGrp:=m.Group("/api/v1/auth")
	authGrp.POST("/register",app.Auth.Register)
	authGrp.POST("/login",app.Auth.Login)
	authGrp.POST("/verify-otp",app.Auth.VerifyOTP)

	m.Start(":8080")
}
