package app

import (
	appconfig "backend/app/config"
	"backend/internal/config"
	serviceconfig "backend/internal/config/services"
	postgres "backend/internal/database/sql/postgres"
	authservice "backend/internal/service/auth"
	"backend/pkg/cryptoutil"
	"backend/pkg/cryptoutil/argon2id"
	"backend/pkg/email"
	"backend/pkg/email/render"
	"backend/pkg/email/resend"
	"backend/pkg/jwt"
	"fmt"
	"net"
	"net/http"

	"github.com/impl0x/go-utils/logger"
)

type App struct {
	t transport
}

// all services
type services struct {
	auth *authservice.Service
}

// all extra utility packages used by services
type pkgs struct {
	otp           cryptoutil.OtpManager
	password      cryptoutil.PasswordHasher
	jwt           jwt.Manager
	emailSender   email.Sender
	emailRenderer *render.Renderer
}

// returns a new configured http client
func newHttpClient(cfg config.HttpClient) *http.Client {
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
func newHttpServer(cfg config.HttpServer) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

// returns all packages with their configured values
func newPkgs(cfg *config.Config) (*pkgs, error) {

	// create new email sender
	httpClient := newHttpClient(cfg.Infra.HttpClient)
	var emailSender email.Sender
	switch appconfig.CurrentEmailSender {
	case appconfig.Resend:
		emailSender = resend.New(httpClient, cfg.Infra.Email.EmailID, cfg.Infra.Email.ResendApiKey)
	default:
		panic("invalid email sender present in app config")
	}

	// create email renderer
	emailRenderer, err := render.New(cfg.App.Name, cfg.App.LogoUrl, cfg.App.HomePageLink)
	if err != nil {
		return nil, fmt.Errorf("App: Email renderer: %w", err)
	}

	return &pkgs{
		otp: cryptoutil.NewOtpManager(cfg.App.Name, *cfg.Services.AuthPkgs.OTP.Map()),
		// we use argon2id as the current password hashing implementation algorithm
		password:      argon2id.New(*cfg.Services.AuthPkgs.Password.Argon2ID.Map()),
		jwt:           jwt.NewManager(cfg.Services.AuthPkgs.JWT.Secret),
		emailSender:   emailSender,
		emailRenderer: emailRenderer,
	}, nil
}

// returns services injected with their respective configs, repositories, packages and logger.
func newServices(cfg serviceconfig.Auth, log logger.Logger, r repositories, p *pkgs) services {
	return services{
		auth: authservice.NewService(authservice.Dependencies{
			Log:    log,
			Config: *cfg.Map(),
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

// Creates a new app with the configuration provided
func New(cfg *config.Config) (*App, error) {
	// logger
	log := logger.New(logger.Level(cfg.Log.Int()))

	// pkgs
	p, err := newPkgs(cfg)
	if err != nil {
		return nil, err
	}

	// repositories
	var r repositories
	switch appconfig.CurrentDatabase {
	case appconfig.Postgres:
		pg, err := postgres.New(cfg.Infra.Postgres, log)
		if err != nil {
			return nil, fmt.Errorf("App: instantiate postgres: %w", err)
		}
		r = newRepositoriesFromPostgres(pg)
	default:
		panic("invalid database present in app config")
	}

	// services
	s := newServices(cfg.Services.Auth, log, r, p)

	// transport
	var t transport
	switch appconfig.CurrentTransport {
	case appconfig.HTTP:
		t = newHttpTransport(log, s, cfg.Infra.HttpServer)
	default:
		panic("invalid transport present in app config")
	}

	return &App{
		t: t,
	}, nil
}

func (a *App) Run() error {
	return a.t.start()
}
