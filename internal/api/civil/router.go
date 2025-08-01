package civil

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func CivilRouter(log *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()

	r.Get("/", GetAllRecords(log))
	return r
}
