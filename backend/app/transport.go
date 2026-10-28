package app

import (
	"backend/internal/config"
	httpserver "backend/internal/transport/http"
	authhandler "backend/internal/transport/http/handler/auth"
	validationhandler "backend/internal/transport/http/handler/validation"
	"backend/internal/transport/http/middleware"
	"backend/pkg/cryptoutil"
	"net/http"

	"github.com/impl0x/go-utils/logger"
	"github.com/impl0x/mo"
)

type transport interface {
	// starts listening for new connections to serve
	start() error
}

// transport using http based on mo framework
type httpTransport struct {
	server *http.Server
	mo     *mo.Mo
	log    logger.Logger
}

type httpHandlers struct {
	auth        authhandler.Handler
	validations validationhandler.Handler
}

// returns a configured Mo api and registers paths, middlewares, etc for handlers
func newHttpTransport(log logger.Logger, s services, cfg config.HttpServer) httpTransport {
	// create handlers
	handlers := httpHandlers{
		auth:        authhandler.NewHandler(s.auth),
		validations: validationhandler.NewHandler(),
	}

	// add validations
	httpserver.AddValidations()

	// create new Mo instance
	m := mo.New()

	// add error handler
	m.HTTPErrorHandler = httpserver.ErrorHandler(log)

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
	handlers.auth.RegisterPaths(v1.Group("/auth"))
	handlers.validations.RegisterPaths(v1.Group("/validation"))

	// create http transport and return
	a := httpTransport{
		server: newHttpServer(cfg),
		mo:     m,
	}
	return a
}

func (h httpTransport) start() error {
	h.log.Info("App running on address " + h.server.Addr)
	return h.mo.StartServer(h.server)
}
