package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eif-courses/civilregistry/internal/api"
	"github.com/eif-courses/civilregistry/internal/config"
	"github.com/eif-courses/civilregistry/internal/logger"
	"github.com/eif-courses/civilregistry/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := logger.NewLogger()
	defer log.Sync()

	cfg := config.Load()

	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalw("Failed to create connection pool", "error", err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalw("Failed to ping database", "error", err)
	}

	queries := repository.New(dbpool)
	router := api.NewRouter(queries, log)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Infow("Starting server", "port", cfg.Port, "database", "connected")

	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatalw("Server failed", "error", err)
	}
}
