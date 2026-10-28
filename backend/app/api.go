package app

import (
	"backend/internal/config"
	"backend/internal/restapi"
	"backend/internal/restapi/middleware"
	"backend/pkg/cryptoutil"
	"net/http"

	"github.com/impl0x/go-utils/logger"
	"github.com/impl0x/mo"
)

type api interface {
	start() error
}

type moHttpApi struct {
	server *http.Server
	mo     *mo.Mo
}

// returns a configured Mo api and registers paths, middlewares, etc for handlers
func newMoHttpApi(log logger.Logger, h handlers, cfg config.HttpServer) moHttpApi {
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

	a := moHttpApi{
		server: newHTTPServer(cfg),
		mo:     m,
	}

	return a
}

func (h moHttpApi) start() error {
	return h.mo.StartServer(h.server)
}
