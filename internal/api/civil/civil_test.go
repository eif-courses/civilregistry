package civil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetAllRecords(t *testing.T) {
	log := zap.NewNop().Sugar()
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := GetAllRecords(log)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "All civil records", rr.Body.String())
}

func TestCivilRouter(t *testing.T) {
	log := zap.NewNop().Sugar()
	router := CivilRouter(log)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "All civil records", rr.Body.String())
}
