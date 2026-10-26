package app

import (
	"backend/internal/config"
	"backend/internal/database/postgres"
	"backend/internal/repository"
	"backend/internal/restapi"
	authhandler "backend/internal/restapi/handler/auth"
	validationhandler "backend/internal/restapi/handler/validation"
	"backend/internal/restapi/middleware"
	authservice "backend/internal/service/auth"
	"backend/internal/validation"
	"backend/pkg/cryptoutil"
	"backend/pkg/email"
	"backend/pkg/email/render"
	"backend/pkg/email/resend"
	"backend/pkg/jwt"
	"fmt"
	"net"
	"net/http"

	"github.com/impl0x/go-utils/logger"
	"github.com/impl0x/mo"
)

type App struct {
	api    *mo.Mo
	server *http.Server
	log    logger.Logger
}

// api handlers
type handlers struct {
	auth        authhandler.Handler
	validations validationhandler.Handler
}

// all services
type services struct {
	auth *authservice.Service
}

// repositories used by the services
type repositories struct {
	user        repository.UserRepository
	userSession repository.UserSessionRepository
	profile     repository.ProfileRepository
}

// all extra utility packages used by services
type pkgs struct {
	otp           cryptoutil.OtpManager
	password      cryptoutil.PasswordManager
	jwt           jwt.Manager
	emailSender   email.Sender
	emailRenderer *render.Renderer
	postgres      postgres.Postgres
}

// returns a new configured http client
func newHTTP(cfg config.HTTP) *http.Client {
	dialer := &net.Dialer{
		Timeout:   cfg.DialerTimeout,
		KeepAlive: cfg.DialerKeepAlive,
	}
	return &http.Client{
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
			ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
			ExpectContinueTimeout: cfg.ExpectContinueTimeout,

			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
			MaxConnsPerHost:     cfg.MaxConnsPerHost,
			IdleConnTimeout:     cfg.IdleConnTimeout,
			ForceAttemptHTTP2:   cfg.ForceAttemptHTTP2,
		},
		Timeout: cfg.Timeout,
	}
}

// returns a new configured http server
func newHTTPServer(cfg config.RestAPI) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

// returns all packages with their configured values
func newPkgs(cfg *config.Config, log logger.Logger) (*pkgs, error) {
	// create database connection
	pg, err := postgres.New(cfg.Infra.Postgres, log)
	if err != nil {
		return nil, fmt.Errorf("App: Could not instantiate postgres, %w", err)
	}
	// create email renderer
	emailRenderer, err := render.New(cfg.App.Name, cfg.App.LogoUrl, cfg.App.HomePageLink)
	if err != nil {
		return nil, fmt.Errorf("App: Email renderer fail, %w", err)
	}
	// create a email sender, for that we are using the Resend API as of now
	httpClient := newHTTP(cfg.Infra.HTTP)
	emailSender := resend.New(httpClient, cfg.Infra.Email.EmailID, cfg.Infra.Email.ResendApiKey)
	return &pkgs{
		otp:           cryptoutil.NewOtpGenerator(cfg.App.Name, cfg.Services.Auth.OTP),
		password:      cryptoutil.NewPasswordManager(config.DefaultAuth.Password),
		jwt:           jwt.NewManager(cfg.Services.Auth.JWT.Secret),
		emailSender:   emailSender,
		emailRenderer: emailRenderer,
		postgres:      pg,
	}, nil
}

// returns repositories made using postgres implementations
func newRepositoriesFromPostgres(pg postgres.Postgres) repositories {
	return repositories{
		user:        pg.Users,
		userSession: pg.UserSessions,
		profile:     pg.Profiles,
	}
}

// returns services injected with their respective configs, repositories, packages and logger.
func newServices(cfg config.Services, log logger.Logger, r repositories, p *pkgs) services {
	return services{
		auth: authservice.NewService(authservice.Dependencies{
			Log:    log,
			Config: cfg.Auth.Service,
			Repos: authservice.Repositories{
				User:    r.user,
				Session: r.userSession,
				Profile: r.profile,
			},
			Jwt:           p.jwt,
			Otp:           p.otp,
			Password:      p.password,
			EmailSender:   p.emailSender,
			EmailRenderer: p.emailRenderer,
		}),
	}
}

// returns handlers injected with services
func newHandlers(s services) handlers {
	return handlers{
		auth:        authhandler.NewHandler(s.auth),
		validations: validationhandler.NewHandler(),
	}
}

// returns a configured Mo api and registers paths, middlewares, etc for handlers
func newMoApi(log logger.Logger, h handlers) *mo.Mo {
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

// Creates a new app with the configuration provided
func New(cfg *config.Config) (*App, error) {
	// logger
	log := logger.New(logger.Level(cfg.Log.Int()), logger.WithOutputBatchingOFF)

	// pkgs
	p, err := newPkgs(cfg, log)
	if err != nil {
		return nil, err
	}

	// repositories
	r := newRepositoriesFromPostgres(p.postgres)

	// services
	s := newServices(cfg.Services, log, r, p)

	// handlers
	h := newHandlers(s)

	// validations
	validation.AddValidations()

	// api - using mo package as http api framework
	api := newMoApi(log, h)

	// server
	server := newHTTPServer(cfg.Infra.RestAPI)

	return &App{
		api:    api,
		server: server,
		log:    log,
	}, nil
}

func (a *App) Run() error {
	a.log.Info("App running on address " + a.server.Addr)
	return a.api.StartServer(a.server)
}
