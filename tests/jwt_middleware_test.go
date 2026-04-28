package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

func TestJWTAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "jwt-test-secret"
	t.Setenv("JWT_SECRET", secret)

	t.Run("missing authorization header", func(t *testing.T) {
		router := gin.New()
		router.GET("/secure", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		router := gin.New()
		router.GET("/secure", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		router := gin.New()
		router.GET("/secure", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		claims := jwt.MapClaims{
			"tenant_id": uuid.New().String(),
			"exp":       time.Now().Add(-time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		req.Header.Set("Authorization", "Bearer "+signed)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		tenantID := uuid.New()
		router := gin.New()
		router.GET("/secure", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
			value, ok := c.Get(middleware.TenantIDContextKey)
			if !ok {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if value.(uuid.UUID) != tenantID {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Status(http.StatusOK)
		})

		claims := jwt.MapClaims{
			"tenant_id": tenantID.String(),
			"exp":       time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		req.Header.Set("Authorization", "Bearer "+signed)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.Code)
		}
	})
}
