package main

import (
	"backend/internal/config"
	"backend/internal/features/auth"
	"backend/internal/utils"
	"context"
	"log"

	"github.com/impl0x/mo"
	"github.com/impl0x/mo/middlewares"
	"github.com/jackc/pgx/v5/pgxpool"
)

// to change later
func main() {
	// println(config.DbConnStr)
	m := mo.New()
	m.HTTPErrorHandler = utils.CustomErrorHandler
	m.AddPostMiddleware(middlewares.LoggerWithResponseCode)
	db, err := pgxpool.New(context.Background(), config.DbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	auth.Register(m.Group("/api/v1/auth"), db)
	m.Start(":8080")
}
