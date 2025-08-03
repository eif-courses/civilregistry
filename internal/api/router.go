package api

import (
	"github.com/eif-courses/civilregistry/internal/api/civil"
	"github.com/eif-courses/civilregistry/internal/repository"
	"github.com/eif-courses/civilregistry/internal/web"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
)

func NewRouter(queries *repository.Queries, log *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()

	// Mount API subrouter
	r.Mount("/api/civil", civil.CivilRouter(queries, log))

	// Add web routes - IMPORTANT: Add this!
	web.SetupRoutes(r, queries, log)

	// Serve assets
	workDir, _ := filepath.Abs(".")
	assetsDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(r, "/assets", assetsDir)

	return r
}

// FileServer stays the same...
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
