package main

import (
	"backend/app"
	"backend/internal/config"
	"fmt"
)

func main() {
	err := startApp()
	if err != nil {
		println("Could not start app", err.Error())
	}
}

func startApp() error {
	// load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("Error loading config, %w", err)
	}

	// create new app
	a, err := app.NewApp(cfg)
	if err != nil {
		return fmt.Errorf("Error instantiating new app, %w", err)
	}

	// run the app
	err = a.Run()
	if err != nil {
		return fmt.Errorf("Error running app, %w", err)
	}
	return nil
}
