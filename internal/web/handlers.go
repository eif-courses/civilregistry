package web

import (
	"net/http"

	"github.com/eif-courses/civilregistry/components/frontend" // Make sure this matches your folder structure
	"go.uber.org/zap"
)

type Handlers struct {
	service *Service
	logger  *zap.SugaredLogger
}

func NewHandlers(service *Service, logger *zap.SugaredLogger) *Handlers {
	return &Handlers{
		service: service,
		logger:  logger,
	}
}

func (h *Handlers) HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	component := frontend.HomePage()
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

	component := frontend.PostsPage(posts)
	if err := component.Render(r.Context(), w); err != nil {
		h.logger.Errorf("Failed to render posts page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
