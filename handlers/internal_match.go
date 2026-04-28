package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/realtime"
)

type matchFoundRequest struct {
	TenantID        string  `json:"tenant_id" binding:"required"`
	PirateURL       string  `json:"pirate_url" binding:"required"`
	ConfidenceScore float64 `json:"confidence_score" binding:"required"`
}

// InternalMatchFoundHandler ingests match detections from async workers.
func InternalMatchFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		gormDB := db.GetDB()
		if gormDB == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
			return
		}

		var req matchFoundRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		tenantID, err := uuid.Parse(strings.TrimSpace(req.TenantID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
			return
		}

		pirateURL := strings.TrimSpace(req.PirateURL)
		if pirateURL == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid pirate_url"})
			return
		}

		if req.ConfidenceScore < 0 || req.ConfidenceScore > 1 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid confidence_score"})
			return
		}

		detection := db.PiracyDetection{
			TenantID:        tenantID,
			PirateURL:       pirateURL,
			ConfidenceScore: req.ConfidenceScore,
			CreatedAt:       time.Now().UTC(),
		}

		if err := gormDB.Create(&detection).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to store detection"})
			return
		}

		_ = realtime.Publish(tenantID, renderDetectionSnippet(detection))
		c.Status(http.StatusOK)
	}
}

func renderDetectionSnippet(detection db.PiracyDetection) string {
	escapedURL := template.HTMLEscapeString(detection.PirateURL)
	escapedScore := template.HTMLEscapeString(fmt.Sprintf("%.2f", detection.ConfidenceScore))
	elapsed := template.HTMLEscapeString(time.Since(detection.CreatedAt).Round(time.Second).String())

	return "<div class=\"rounded-md border border-rose-200 bg-rose-50 p-3\">" +
		"<div class=\"text-xs uppercase text-rose-600\">Piracy Match</div>" +
		"<div class=\"mt-2 text-sm font-semibold text-rose-900\">" + escapedURL + "</div>" +
		"<div class=\"mt-1 text-xs text-rose-700\">Confidence: " + escapedScore + "</div>" +
		"<div class=\"mt-1 text-xs text-rose-600\">Detected: " + elapsed + " ago</div>" +
		"</div>"
}
