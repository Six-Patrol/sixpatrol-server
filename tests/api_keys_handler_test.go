package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/handlers"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

func TestGenerateAPIKeyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gormDB := setupTestDB(t)

	t.Run("missing tenant", func(t *testing.T) {
		router := gin.New()
		router.POST("/keys", handlers.GenerateAPIKeyHandler())

		req := httptest.NewRequest(http.MethodPost, "/keys", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("valid tenant", func(t *testing.T) {
		tenantID := uuid.New()
		router := gin.New()
		router.POST("/keys", func(c *gin.Context) {
			c.Set(middleware.TenantIDContextKey, tenantID)
			handlers.GenerateAPIKeyHandler()(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/keys", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.Code)
		}

		var apiKey db.ApiKey
		if err := gormDB.Order("created_at desc").First(&apiKey).Error; err != nil {
			t.Fatalf("failed to fetch api key: %v", err)
		}
		if apiKey.ApiKeyString == "" || apiKey.SecretKeyString == "" {
			t.Fatalf("expected api key and secret key to be populated")
		}
		if !strings.Contains(resp.Body.String(), apiKey.SecretKeyString) {
			t.Fatalf("response should include secret key")
		}
		if !strings.Contains(resp.Body.String(), apiKey.ApiKeyString) {
			t.Fatalf("response should include api key")
		}
	})
}
