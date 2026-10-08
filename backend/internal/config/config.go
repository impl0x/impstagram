package config

import (
	"backend/pkg/env"
	"backend/pkg/structutils"
	"fmt"
	"time"

	"github.com/impl0x/mo/validator/v3"
)

const _APP_VERSION = "0.5.0"

type Config struct {
	App      app
	Services Services
	Infra    Infrastructure
	Log      log
}

type Services struct {
	Auth AuthConfig
	Http HttpConfig // http requests client
}
type Infrastructure struct {
	RestAPI  restApi
	Postgres postgres
	Email    email
}

type app struct {
	Name    string `env:"APP_NAME,required"`
	Version string // to be set by the program, as version changes by each compile not by deployment
}

type restApi struct {
	Port              string        `env:"REST_API_PORT,required"`
	ReadTimeout       time.Duration `env:"REST_API_READ_TIMEOUT,default=5s"`
	ReadHeaderTimeout time.Duration `env:"REST_API_READ_HEADER_TIMEOUT,default=2s"`
	WriteTimeout      time.Duration `env:"REST_API_WRITE_TIMEOUT,default=10s"`
	IdleTimeout       time.Duration `env:"REST_API_IDLE_TIMEOUT,default=120s"`
}

type postgres struct {
	PoolMax int    `env:"PG_POOL_MAX" validate:"gte=0"`
	URL     string `env:"PG_URL,required" validate:"url"`
}

type email struct {
	ResendApiKey string `env:"EMAIL_RESEND_API_KEY,required" validate:"startswith=re_"`
	EmailID      string `env:"EMAIL_ID,required" validate:"email"`
}

type log struct {
	Level string `env:"LOG_LEVEL,required,default=ERROR"`
}

// Loads configs.
func Load() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config parsing error: %w", err)
	}
	err:=validator.Validate(&cfg)
	if err!=nil{
		return nil, fmt.Errorf("config: validation fail for config, %w", err)
	}
	
	// app
	cfg.App.Version = _APP_VERSION

	// services
	// auth
	err = structutils.CopyFields(&cfg.Services.Auth, &DefaultAuthConfig)
	if err != nil {
		panic(fmt.Errorf("util.CopyFields: unexpected error returned, %w", err))
	}
	// http
	err = structutils.CopyFields(&cfg.Services.Http, &DefaultHttpConfig)
	if err != nil {
		panic(fmt.Errorf("util.CopyFields: unexpected error returned, %w", err))
	}

	return &cfg, nil
}
