package post

import (
	restapi "github.com/eif-courses/civilregistry/internal/generated/api/post"
	"github.com/eif-courses/civilregistry/internal/generated/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func SetupRoutes(r chi.Router, queries *repository.Queries, log *zap.SugaredLogger) {
	service := restapi.NewService(queries, log)
	handlers := NewHandlers(service, log)

	// Web routes
	r.Get("/", handlers.HomePage)
	r.Get("/posts", handlers.PostsPage)
}
