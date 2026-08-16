package main

import (
	"backend/app"
	"backend/internal/utils"

	"github.com/impl0x/mo"
	"github.com/impl0x/mo/middlewares"
)

func main() {
	m:=mo.New()
	m.HTTPErrorHandler=utils.CustomErrorHandler
	m.AddPostMiddleware(middlewares.LoggerWithResponseCode)
	app:=app.NewApp(app.NewAuth())
	
	authGrp:=m.Group("/api/v1/auth")
	authGrp.POST("/register",app.Auth.Register)
	authGrp.POST("/login",app.Auth.Login)
	authGrp.POST("/verify-otp",app.Auth.VerifyOTP)

	m.Start(":8080")
}
