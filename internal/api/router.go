package api

import (
	"net/http"
	"path/filepath"

	"github.com/eif-courses/civilregistry/internal/generated/api/post"
	"github.com/eif-courses/civilregistry/internal/generated/repository"
	frontendpost "github.com/eif-courses/civilregistry/internal/web/post"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

func NewRouter(queries *repository.Queries, log *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Swagger documentation route
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Mount("/post", post.PostRouter(queries, log))

		// FORCE REFERENCE: This ensures Swagger sees the handlers
		_ = post.NewHandlers
	})

	// Web routes
	frontendpost.SetupRoutes(r, queries, log)

	// Serve assets
	workDir, _ := filepath.Abs(".")
	assetsDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(r, "/assets", assetsDir)

	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if path == "" {
		panic("FileServer: empty path")
	}

	if path[len(path)-1] != '/' {
		path += "/"
	}

	fs := http.StripPrefix(path, http.FileServer(root))
	r.Get(path+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
