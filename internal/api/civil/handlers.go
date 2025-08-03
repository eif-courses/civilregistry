// internal/api/civil/handlers.go
package civil

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
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

func (h *Handlers) GetPublicPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.service.GetPublicPosts(r.Context())
	if err != nil {
		h.logger.Errorf("Handler error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  posts,
		"count": len(posts),
	})
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.service.HealthCheck(r.Context())
	if err != nil {
		http.Error(w, "Service unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "civilregistry-api",
		"version": "1.0.0",
	})
}
