package civil

import (
	"net/http"

	"go.uber.org/zap"
)

func GetAllRecords(log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infow("Received request", "method", r.Method, "path", r.URL.Path)
		w.Write([]byte("All civil records"))
	}
}
