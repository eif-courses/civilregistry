package civil

import (
	"github.com/eif-courses/civilregistry/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func CivilRouter(queries *repository.Queries, log *zap.SugaredLogger) chi.Router {
	r := chi.NewRouter()

	// Create service with repository
	service := NewService(queries, log)
	handlers := NewHandlers(service, log)

	r.Get("/health", handlers.HealthCheck)
	r.Get("/posts", handlers.GetPublicPosts)

	return r
}
