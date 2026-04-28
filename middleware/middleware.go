package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
)

const TenantIDContextKey = "tenant_id"

// HMACAuthMiddleware validates HMAC signatures for ingest traffic.
func HMACAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := strings.TrimSpace(c.GetHeader("x-api-key"))
		signature := strings.TrimSpace(c.GetHeader("x-signature"))
		if apiKey == "" || signature == "" {
			abortUnauthorized(c)
			return
		}

		gormDB := db.GetDB()
		if gormDB == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
			return
		}

		var apiKeyRecord db.ApiKey
		if err := gormDB.Where("api_key_string = ?", apiKey).First(&apiKeyRecord).Error; err != nil {
			abortUnauthorized(c)
			return
		}

		bodyBytes, err := readAndRestoreBody(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		expected := computeHMACSHA256Hex(bodyBytes, apiKeyRecord.SecretKeyString)
		provided := normalizeSignature(signature)
		if subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) != 1 {
			abortUnauthorized(c)
			return
		}

		c.Set(TenantIDContextKey, apiKeyRecord.TenantID)
		c.Next()
	}
}

// JWTAuthMiddleware validates JWTs for dashboard traffic.
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "jwt secret not configured"})
			return
		}

		tokenString := extractBearerToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			abortUnauthorized(c)
			return
		}

		claims := jwt.MapClaims{}
		parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		parsed, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !parsed.Valid {
			abortUnauthorized(c)
			return
		}

		if exp, err := claims.GetExpirationTime(); err == nil && exp != nil {
			if time.Now().After(exp.Time) {
				abortUnauthorized(c)
				return
			}
		}

		tenantID, err := extractTenantID(claims)
		if err != nil {
			abortUnauthorized(c)
			return
		}

		c.Set(TenantIDContextKey, tenantID)
		c.Next()
	}
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}

func normalizeSignature(signature string) string {
	sig := strings.ToLower(strings.TrimSpace(signature))
	sig = strings.TrimPrefix(sig, "sha256=")
	sig = strings.TrimPrefix(sig, "hmac-sha256=")
	return sig
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func extractTenantID(claims jwt.MapClaims) (uuid.UUID, error) {
	value, ok := claims["tenant_id"]
	if !ok {
		value = claims["sub"]
	}
	valueStr, ok := value.(string)
	if !ok || strings.TrimSpace(valueStr) == "" {
		return uuid.Nil, fmt.Errorf("tenant_id missing")
	}

	return uuid.Parse(valueStr)
}

func computeHMACSHA256Hex(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func readAndRestoreBody(c *gin.Context) ([]byte, error) {
	if c.Request.Body == nil {
		return []byte{}, nil
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	c.Request.ContentLength = int64(len(bodyBytes))
	return bodyBytes, nil
}
