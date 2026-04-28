package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

const (
	apiKeyBytes    = 32
	secretKeyBytes = 48
)

// GenerateAPIKeyHandler creates a new API key/secret pair for the tenant.
func GenerateAPIKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		gormDB := db.GetDB()
		if gormDB == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
			return
		}

		tenantIDValue, ok := c.Get(middleware.TenantIDContextKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "tenant not authorized"})
			return
		}

		tenantID, ok := tenantIDValue.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid tenant"})
			return
		}

		apiKey, err := generateRandomKey(apiKeyBytes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate api key"})
			return
		}

		secretKey, err := generateRandomKey(secretKeyBytes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate secret key"})
			return
		}

		record := db.ApiKey{
			TenantID:        tenantID,
			ApiKeyString:    apiKey,
			SecretKeyString: secretKey,
		}

		if err := gormDB.Create(&record).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to store api key"})
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderAPIKeySnippet(apiKey, secretKey)))
	}
}

func generateRandomKey(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func renderAPIKeySnippet(apiKey, secretKey string) string {
	escapedAPIKey := template.HTMLEscapeString(apiKey)
	escapedSecret := template.HTMLEscapeString(secretKey)
	return "<div class=\"rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900\">" +
		"<div class=\"text-sm font-semibold\">New Secret Key</div>" +
		"<div class=\"mt-2 font-mono text-xs break-all\">" + escapedSecret + "</div>" +
		"<p class=\"mt-2 text-xs text-amber-700\">This secret key will only be shown once. Store it securely.</p>" +
		"<div class=\"mt-3 text-xs text-slate-700\">API Key: <span class=\"font-mono\">" + escapedAPIKey + "</span></div>" +
		"</div>"
}
