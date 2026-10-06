package config

import (
	"fmt"

	"backend/internal/util"
	"backend/pkg/env"
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
}
type Infrastructure struct {
	HTTP     restHttp
	Postgres postgres
	Email    email
}

type app struct {
	Name    string `env:"APP_NAME,required"`
	Version string // to be set by the program, as version changes by each compile not by deployment
}

type restHttp struct {
	Port string `env:"REST_HTTP_PORT,required"`
}

type postgres struct {
	PoolMax int    `env:"PG_POOL_MAX,required" validate:"gte=0"`
	URL     string `env:"PG_URL,required" validate:"url"`
}

type email struct {
	ResendApiKey string `env:"EMAIL_RESEND_API_KEY,required"`
}

type log struct {
	Level string `env:"LOG_LEVEL,required"`
}

// Loads configs.
func Load() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config parsing error: %w", err)
	}
	// app
	cfg.App.Version = _APP_VERSION

	// services
	// auth
	err := util.CopyFields(&cfg.Services.Auth, &defaultAuthConfig)
	if err != nil {
		panic(fmt.Errorf("util.CopyFields unexpected error returned, %w", err))
	}

	return &cfg, nil
}
