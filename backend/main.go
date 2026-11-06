package main

import (
	"backend/app"
	"backend/internal/config"
	"fmt"
)

func main() {
	err := startApp()
	if err != nil {
		println("Error: Could not start app:", err.Error())
	}
}

func startApp() error {
	// load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// create new app
	a, err := app.New(cfg)
	if err != nil {
		return fmt.Errorf("instantiate new app: %w", err)
	}

	// run the app
	err = a.Run()
	if err != nil {
		return fmt.Errorf("run app: %w", err)
	}
	return nil
}
