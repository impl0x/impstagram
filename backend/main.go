package main

import (
	"backend/app"
	"fmt"
)

func main() {
	startApp()
}

func startApp() {
	a, err := app.NewApp()
	if err != nil {
		fmt.Println("Error instantiating new app,", err)
		return
	}
	err = a.Run()
	if err != nil {
		fmt.Println("Error running app, %w", err)
	}
}
