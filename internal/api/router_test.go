package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMainRouter(t *testing.T) {
	log := zap.NewNop().Sugar()
	router := NewRouter(log) // No api. prefix needed because same package

	req, _ := http.NewRequest("GET", "/api/civil/", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "All civil records", rr.Body.String())
}
