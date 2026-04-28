package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/handlers"
	"github.com/sixpatrol/sixpatrol-server/realtime"
)

func TestInternalMatchFoundHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gormDB := setupTestDB(t)

	broadcaster := realtime.NewBroadcaster()
	previous := realtime.SetBroadcaster(broadcaster)
	defer realtime.SetBroadcaster(previous)

	tenantID := uuid.New()
	msgCh, cancel := realtime.Subscribe(tenantID)
	defer cancel()

	router := gin.New()
	router.POST("/internal", handlers.InternalMatchFoundHandler())

	t.Run("invalid payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/internal", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.Code)
		}
	})

	t.Run("invalid tenant_id", func(t *testing.T) {
		payload := map[string]any{
			"tenant_id":        "not-a-uuid",
			"pirate_url":       "https://pirate.example/video",
			"confidence_score": 0.98,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/internal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.Code)
		}
	})

	t.Run("invalid confidence_score", func(t *testing.T) {
		payload := map[string]any{
			"tenant_id":        tenantID.String(),
			"pirate_url":       "https://pirate.example/video",
			"confidence_score": 1.5,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/internal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		payload := map[string]any{
			"tenant_id":        tenantID.String(),
			"pirate_url":       "https://pirate.example/video",
			"confidence_score": 0.98,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/internal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.Code)
		}

		var count int64
		if err := gormDB.Model(&db.PiracyDetection{}).
			Where("tenant_id = ?", tenantID.String()).
			Count(&count).Error; err != nil {
			t.Fatalf("failed to count detections: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected 1 detection, got %d", count)
		}

		select {
		case msg := <-msgCh:
			if !strings.Contains(msg, "Piracy Match") {
				t.Fatalf("expected message to include Piracy Match")
			}
			if !strings.Contains(msg, "pirate.example") {
				t.Fatalf("expected message to include pirate URL")
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("expected broadcast message")
		}
	})
}
