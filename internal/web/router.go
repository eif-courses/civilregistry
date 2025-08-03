package web

import (
	"github.com/eif-courses/civilregistry/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func SetupRoutes(r chi.Router, queries *repository.Queries, log *zap.SugaredLogger) {
	service := NewService(queries, log)
	handlers := NewHandlers(service, log)

	// Web routes
	r.Get("/", handlers.HomePage)
	r.Get("/posts", handlers.PostsPage)
}
