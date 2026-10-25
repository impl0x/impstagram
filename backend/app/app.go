package app

import (
	"backend/internal/config"
	"backend/internal/database/postgres"
	dbpostgres "backend/internal/database/postgres"
	"backend/internal/repository"
	"backend/internal/restapi"
	authhandler "backend/internal/restapi/handler/auth"
	validationhandler "backend/internal/restapi/handler/validation"
	"backend/internal/restapi/middleware"
	authservice "backend/internal/service/auth"
	"backend/internal/validation"
	"backend/pkg/cryptoutil"
	"backend/pkg/email"
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

type handlers struct {
	auth        authhandler.Handler
	validations validationhandler.Handler
}

type services struct {
	auth *authservice.Service
}

type repositories struct {
	user        repository.UserRepository
	userSession repository.UserSessionRepository
	profile     repository.ProfileRepository
}

type pkgs struct {
	otp      cryptoutil.OtpManager
	password cryptoutil.PasswordManager
	jwt      jwt.Manager
	email    email.Manager
	postgres postgres.Postgres
}

// extra initialization functions

func newHTTP(cfg config.HTTP) *http.Client {
	dialer := &net.Dialer{
		Timeout:   cfg.DialerTimeout,
		KeepAlive: cfg.DialerKeepAlive,
	}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,

		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		ForceAttemptHTTP2:   cfg.ForceAttemptHTTP2,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}

func newHTTPServer(cfg config.RestAPI) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

func NewApp() (*App, error) {
	// config
	conf, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("App: Config load error, %w", err)
	}

	// logger
	log := logger.New(logger.Level(conf.Log.Int()), logger.WithOutputBatchingOFF)

	// Infra - http, email, postgres
	httpClient := newHTTP(conf.Infra.HTTP)
	emailRenderer, err := email.NewRenderer(conf.App.Name, conf.App.LogoUrl, conf.App.HomePageLink)
	emailManager := email.NewManager(conf.Infra.Email, httpClient, emailRenderer)
	if err != nil {
		return nil, fmt.Errorf("App: Email renderer fail, %w", err)
	}
	pg, err := dbpostgres.NewPostgres(conf.Infra.Postgres, log)
	if err != nil {
		return nil, fmt.Errorf("App: Could not instantiate postgres, %w", err)
	}

	// pkgs
	p := pkgs{
		otp:      cryptoutil.NewOtpGenerator(conf.App.Name, conf.Services.Auth.OTP),
		password: cryptoutil.NewPasswordManager(config.DefaultAuth.Password),
		jwt:      jwt.NewManager(conf.Services.Auth.JWT.Secret),
		email:    emailManager,
		postgres: pg,
	}

	// repositories
	d := repositories{
		user:        pg.Users,
		userSession: pg.UserSessions,
		profile:     pg.Profiles,
	}

	// services
	s := services{
		auth: authservice.NewService(authservice.Dependencies{
			Log:    log,
			Config: conf.Services.Auth.Service,
			Repos: authservice.Repositories{
				User:    d.user,
				Session: d.userSession,
				Profile: d.profile,
			},
			Jwt:      p.jwt,
			Otp:      p.otp,
			Password: p.password,
			Email:    p.email,
		}),
	}

	h := handlers{
		auth:        authhandler.NewHandler(s.auth),
		validations: validationhandler.NewHandler(),
	}

	// validations
	validation.AddValidations()

	// api
	m := mo.New()
	{
		// add middlewares and other stuff
		m.HTTPErrorHandler = restapi.ErrorHandler(log)
		m.AddPreMiddleware(middleware.AddRequestID(cryptoutil.GeneratePseudoHexToken64Bit))
		m.AddPreMiddleware(middleware.AddStartTime())
		m.AddPostMiddleware(middleware.RequestLogger(log))
		m.Use(middleware.Recover(true, log))
		m.Config.PrintStartMsg = false
	}

	{
		// register groups and paths
		v1 := m.Group("/v1")
		h.auth.RegisterPaths(v1.Group("/auth"))
		h.validations.RegisterPaths(v1.Group("/validation"))
	}

	// server
	server := newHTTPServer(conf.Infra.RestAPI)

	log.Debug("New App successfully created")
	return &App{
		api:    m,
		server: server,
		log:    log,
	}, nil
}

func (a *App) Run() error {
	a.log.Info("App running on address " + a.server.Addr)
	return a.api.StartServer(a.server)
}
