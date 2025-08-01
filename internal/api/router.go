package api

import (
	"net/http"

	"github.com/eif-courses/civilregistry/internal/api/civil"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewRouter(log *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()

	// pass logger to subrouters
	r.Mount("/api/civil", civil.CivilRouter(log))

	return r
}
