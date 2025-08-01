package main

import (
	"net/http"

	"github.com/eif-courses/civilregistry/internal/api"
	"github.com/eif-courses/civilregistry/internal/logger"
)

func main() {
	log := logger.NewLogger()
	defer log.Sync() // flush logs

	router := api.NewRouter(log)

	log.Infow("Starting server", "port", 8080)
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalw("Server failed", "error", err)
	}
}
