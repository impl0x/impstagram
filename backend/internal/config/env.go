package config

type environment = string

const (
	devEnv  environment = "dev"
	prodEnv environment = "prod"
)

const Env = devEnv // switch when env is prod
