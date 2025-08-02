package api

import (
	"net/http"
	"path/filepath"

	"github.com/eif-courses/civilregistry/internal/api/civil"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewRouter(log *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()

	// Mount API subrouter
	r.Mount("/api/civil", civil.CivilRouter(log))

	// Serve /assets/* (JS, CSS, etc. for templ, tailwind)
	workDir, _ := filepath.Abs(".")
	assetsDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(r, "/assets", assetsDir)

	return r
}

// FileServer sets up a handler to serve static files under a given path.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if path == "" {
		panic("FileServer: empty path")
	}

	// Ensure path ends with "/"
	if path[len(path)-1] != '/' {
		path += "/"
	}

	fs := http.StripPrefix(path, http.FileServer(root))
	r.Get(path+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
