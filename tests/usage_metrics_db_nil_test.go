package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/handlers"
)

func TestUsageMetricsHandler_DBNotInitialized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db.SetDB(nil)
	defer db.SetDB(nil)

	router := gin.New()
	router.GET("/usage", handlers.UsageMetricsHandler())

	req := httptest.NewRequest(http.MethodGet, "/usage", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}
