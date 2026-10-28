package app

import (
	"backend/internal/restapi"
	"backend/internal/restapi/middleware"
	"backend/pkg/cryptoutil"

	"github.com/impl0x/go-utils/logger"
	"github.com/impl0x/mo"
)

type api interface {
	start() error
}

type moRestApi struct{
	
}

// returns a configured Mo api and registers paths, middlewares, etc for handlers
func newMoRestApi(log logger.Logger, h handlers) *mo.Mo {
	// create new Mo instance
	m := mo.New()

	// add error handler
	m.HTTPErrorHandler = restapi.ErrorHandler(log)

	// add pre and post middlewares
	m.AddPreMiddleware(middleware.AddRequestID(cryptoutil.GeneratePseudoHexToken64Bit))
	m.AddPreMiddleware(middleware.AddStartTime())
	m.AddPostMiddleware(middleware.RequestLogger(log))

	// add middlewares
	m.Use(middleware.Recover(true, log))

	// extra configurations
	m.Config.PrintStartMsg = false

	// register groups and paths for handlers
	v1 := m.Group("/v1")
	h.auth.RegisterPaths(v1.Group("/auth"))
	h.validations.RegisterPaths(v1.Group("/validation"))

	return m
}
