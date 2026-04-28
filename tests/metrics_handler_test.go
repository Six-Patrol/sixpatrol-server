package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/handlers"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

func TestUsageMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gormDB := setupTestDB(t)

	tenantID := uuid.New()
	otherTenant := uuid.New()
	now := time.Now().UTC()

	rows := []db.TelemetryUsage{
		{TenantID: tenantID, FeatureUsed: "dino", FramesProcessed: 100, Timestamp: now.Add(-24 * time.Hour)},
		{TenantID: tenantID, FeatureUsed: "dino", FramesProcessed: 200, Timestamp: now.Add(-10 * time.Hour)},
		{TenantID: tenantID, FeatureUsed: "dino", FramesProcessed: 300, Timestamp: now.Add(-45 * 24 * time.Hour)},
		{TenantID: otherTenant, FeatureUsed: "dino", FramesProcessed: 500, Timestamp: now.Add(-24 * time.Hour)},
	}
	for _, row := range rows {
		if err := gormDB.Create(&row).Error; err != nil {
			t.Fatalf("failed to insert telemetry row: %v", err)
		}
	}

	router := gin.New()
	router.GET("/usage", func(c *gin.Context) {
		c.Set(middleware.TenantIDContextKey, tenantID)
		handlers.UsageMetricsHandler()(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/usage", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	body := resp.Body.String()
	if !strings.Contains(body, ">300<") {
		t.Fatalf("expected total frames 300 in response")
	}
	if !strings.Contains(body, "$0.30") {
		t.Fatalf("expected bill $0.30 in response")
	}
}
