package app

import (
	"backend/internal/config"
	"backend/internal/database/postgres"

	authservice "backend/internal/service/auth"
	"backend/pkg/cryptoutil"
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
	password      cryptoutil.PasswordManager
	jwt           jwt.Manager
	emailSender   email.Sender
	emailRenderer *render.Renderer
	postgres      postgres.Postgres
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
	httpClient := newHttpClient(cfg.Infra.HttpClient)
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

	// transport
	t := newHttpTransport(log, s, cfg.Infra.HttpServer)

	return &App{
		t: t,
	}, nil
}

func (a *App) Run() error {
	return a.t.start()
}
