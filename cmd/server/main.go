package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eif-courses/civilregistry/internal/api"
	"github.com/eif-courses/civilregistry/internal/config"
	"github.com/eif-courses/civilregistry/internal/generated/repository"
	"github.com/eif-courses/civilregistry/internal/logger"

	// Import generated swagger docs
	_ "github.com/eif-courses/civilregistry/docs"
	_ "github.com/eif-courses/civilregistry/internal/generated/api/post"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @title Civil Registry API
// @version 1.0
// @description This is the Civil Registry API server with auto-generated documentation.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

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
	log.Infow("Starting server",
		"port", cfg.Port,
		"database", "connected",
		"swagger", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Port),
	)

	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatalw("Server failed", "error", err)
	}
}
