package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/handlers"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

func TestTelemetryIngestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gormDB := setupTestDB(t)

	t.Run("missing tenant", func(t *testing.T) {
		router := gin.New()
		router.POST("/telemetry", handlers.TelemetryIngestHandler())

		payload := []byte(`{"action":"dino_v2_inference","frames_processed":10,"timestamp":123}`)
		req := httptest.NewRequest(http.MethodPost, "/telemetry", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		tenantID := uuid.New()
		router := gin.New()
		router.POST("/telemetry", func(c *gin.Context) {
			c.Set(middleware.TenantIDContextKey, tenantID)
			handlers.TelemetryIngestHandler()(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/telemetry", bytes.NewReader([]byte("not-json")))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.Code)
		}
	})

	t.Run("valid payload", func(t *testing.T) {
		tenantID := uuid.New()
		router := gin.New()
		router.POST("/telemetry", func(c *gin.Context) {
			c.Set(middleware.TenantIDContextKey, tenantID)
			handlers.TelemetryIngestHandler()(c)
		})

		payload := map[string]any{
			"action":           "dino_v2_inference",
			"frames_processed": 500,
			"timestamp":        time.Now().Unix(),
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/telemetry", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.Code)
		}

		var count int64
		if err := gormDB.Model(&db.TelemetryUsage{}).Where("tenant_id = ?", tenantID.String()).Count(&count).Error; err != nil {
			t.Fatalf("failed to count telemetry rows: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected 1 telemetry row, got %d", count)
		}
	})
}
