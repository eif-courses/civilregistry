package post

import (
	"net/http"

	restapi "github.com/eif-courses/civilregistry/internal/generated/api/post"
	"github.com/eif-courses/civilregistry/internal/web/ui"
	"go.uber.org/zap"
)

type Handlers struct {
	service *restapi.Service
	logger  *zap.SugaredLogger
}

func NewHandlers(service *restapi.Service, logger *zap.SugaredLogger) *Handlers {
	return &Handlers{
		service: service,
		logger:  logger,
	}
}

func (h *Handlers) HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	component := ui.HomePage()
	if err := component.Render(r.Context(), w); err != nil {
		h.logger.Errorf("Failed to render home page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handlers) PostsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	posts, err := h.service.GetPublicPosts(r.Context())
	if err != nil {
		h.logger.Errorf("Failed to get posts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	component := ui.PostsPage(posts)
	if err := component.Render(r.Context(), w); err != nil {
		h.logger.Errorf("Failed to render posts page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
