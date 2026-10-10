package config

import (
	"backend/pkg/env"
	"backend/pkg/structutils"
	"fmt"

	"github.com/impl0x/mo/validator/v3"
)

const _APP_VERSION = "0.5.0"

type Config struct {
	App      App
	Infra    Infrastructure
	Log      Log
	Services Services
}

type App struct {
	Name    string `env:"APP_NAME,required"`
	Version string // to be set by the program, as version changes by each compile not by deployment
}

type Log struct {
	Level string `env:"LOG_LEVEL"`
}

var DefaultLog = Log{
	Level: "ERROR",
}

type Services struct {
	Auth AuthService
}

// wrapper over [structutils.CopyFields] to panic if an error is returned
//
// we panic because the only time this method can return an error is we pass
// the wrong arguments to it, so it is fine to panic here
func copyFields(a, b any) {
	err := structutils.CopyFields(a, b)
	if err != nil {
		panic(fmt.Errorf("util.CopyFields: unexpected error returned, %w", err))
	}
}

// Loads configs.
func Load() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config parsing error: %w", err)
	}
	err := validator.Validate(&cfg)
	if err != nil {
		return nil, fmt.Errorf("config: validation fail for config, %w", err)
	}

	// -+-+- app -+-+-
	cfg.App.Version = _APP_VERSION

	// -+-+- Infra -+-+-
	// http
	copyFields(&cfg.Infra.HTTP, &DefaultHTTP)
	// restAPI
	copyFields(&cfg.Infra.RestAPI, &DefaultRestAPI)
	// postgres
	copyFields(&cfg.Infra.Postgres, &DefaultPostgres)
	
	// -+-+- Services -+-+-
	// auth
	copyFields(&cfg.Services.Auth, &DefaultAuth)

	return &cfg, nil
}
