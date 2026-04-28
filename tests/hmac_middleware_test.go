package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

func TestHMACAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing headers", func(t *testing.T) {
		setupTestDB(t)
		router := gin.New()
		router.POST("/test", middleware.HMACAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("{}"))
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("invalid api key", func(t *testing.T) {
		setupTestDB(t)
		router := gin.New()
		router.POST("/test", middleware.HMACAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("{}"))
		req.Header.Set("x-api-key", "missing")
		req.Header.Set("x-signature", "deadbeef")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		gormDB := setupTestDB(t)
		tenantID := uuid.New()
		apiKey := "api_test"
		secret := "secret_test"
		if err := gormDB.Create(&db.ApiKey{ID: uuid.New(), TenantID: tenantID, ApiKeyString: apiKey, SecretKeyString: secret}).Error; err != nil {
			t.Fatalf("failed to insert api key: %v", err)
		}

		router := gin.New()
		router.POST("/test", middleware.HMACAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("{\"hello\":\"world\"}"))
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("x-signature", "bad-signature")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("valid signature", func(t *testing.T) {
		gormDB := setupTestDB(t)
		tenantID := uuid.New()
		apiKey := "api_valid"
		secret := "secret_valid"
		if err := gormDB.Create(&db.ApiKey{ID: uuid.New(), TenantID: tenantID, ApiKeyString: apiKey, SecretKeyString: secret}).Error; err != nil {
			t.Fatalf("failed to insert api key: %v", err)
		}

		var gotTenant uuid.UUID
		var gotBody string
		router := gin.New()
		router.POST("/test", middleware.HMACAuthMiddleware(), func(c *gin.Context) {
			value, ok := c.Get(middleware.TenantIDContextKey)
			if ok {
				gotTenant = value.(uuid.UUID)
			}
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			gotBody = string(bodyBytes)
			c.Status(http.StatusOK)
		})

		payload := "{\"hello\":\"world\"}"
		signature := computeHMAC(payload, secret)
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(payload))
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("x-signature", signature)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.Code)
		}
		if gotTenant != tenantID {
			t.Fatalf("expected tenant_id %s, got %s", tenantID, gotTenant)
		}
		if gotBody != payload {
			t.Fatalf("expected body %s, got %s", payload, gotBody)
		}
	})
}

func computeHMAC(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
