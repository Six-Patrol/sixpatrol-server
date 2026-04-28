package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

type telemetryRequest struct {
	Action          string `json:"action" binding:"required"`
	FramesProcessed int64  `json:"frames_processed" binding:"required"`
	Timestamp       int64  `json:"timestamp" binding:"required"`
}

// TelemetryIngestHandler handles POST /ingest/v1/telemetry payloads.
func TelemetryIngestHandler() gin.HandlerFunc {
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

		var req telemetryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		usage := db.TelemetryUsage{
			TenantID:        tenantID,
			FeatureUsed:     req.Action,
			FramesProcessed: req.FramesProcessed,
			Timestamp:       time.Unix(req.Timestamp, 0).UTC(),
		}

		if err := gormDB.Create(&usage).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to record telemetry"})
			return
		}

		c.Status(http.StatusOK)
	}
}
