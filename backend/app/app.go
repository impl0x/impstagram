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
	t transporter
}

// all services
type services struct {
	auth *authservice.Service
}

// all extra utility packages used by services
type pkgs struct {
	otp           cryptoutil.OTPManager
	totp          cryptoutil.TOTPManager
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

	// create new password hasher
	var passwordHasher cryptoutil.PasswordHasher
	switch appconfig.Current().Hasher.Password {
	case appconfig.Argon2ID:
		passwordHasher = argon2id.New(*cfg.Services.AuthPkgs.Password.Argon2ID.Map())
	default:
		panic("invalid password hasher present in app config")
	}

	// create new email sender
	httpClient := newHttpClient(cfg.Infra.HttpClient)
	var emailSender email.Sender
	switch appconfig.Current().Extra.EmailSender {
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
		otp:           cryptoutil.NewOTPManager(*cfg.Services.AuthPkgs.OTP.Map()),
		totp:          cryptoutil.NewTOTPManager(*cfg.Services.AuthPkgs.TOTP.Map()),
		password:      passwordHasher,
		jwt:           jwt.NewManager(cfg.Services.AuthPkgs.JWT.Secret),
		emailSender:   emailSender,
		emailRenderer: emailRenderer,
	}, nil
}

// returns services injected with their respective configs, repositories, packages and logger.
func newServices(cfg serviceconfig.Auth, log logger.Logger, r repositories, p *pkgs) services {
	var h cryptoutil.Hasher
	switch appconfig.Current().Hasher.RefreshToken {
	case appconfig.MD5:
		h = cryptoutil.GenerateMD5Hash
	case appconfig.SHA256:
		h = cryptoutil.GenerateSHA256Hash
	case appconfig.Argon2ID:
		panic("argon2id is not supported for hashing refresh tokens")
	default:
		panic("invalid hashing algorithm")
	}
	return services{
		auth: authservice.NewService(authservice.Dependencies{
			Log:    log,
			Config: *cfg.Map(),
			Repos: authservice.Repositories{
				User:    r.user,
				Session: r.userSession,
				Profile: r.profile,
			},
			Jwt:                p.jwt,
			Otp:                p.otp,
			Totp:               p.totp,
			Password:           p.password,
			EmailSender:        p.emailSender,
			EmailRenderer:      p.emailRenderer,
			RefreshTokenHasher: h,
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
	switch appconfig.Current().Database {
	case appconfig.Postgres:
		pg, err := postgres.New(*cfg.Infra.Postgres.Map(), log)
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
	tp := newTransport(
		log,
		s,
		func() string {
			return cryptoutil.GeneratePseudoHexToken(appconfig.Current().Extra.RequestIDBitSize)
		},
	)
	var t transporter
	switch appconfig.Current().Transport {
	case appconfig.HTTP:
		t = tp.http(cfg.Infra.HttpServer)
	default:
		panic("invalid transport present in app config")
	}

	return &App{
		t: t,
	}, nil
}

func (a App) Run() error {
	return a.t.start()
}
