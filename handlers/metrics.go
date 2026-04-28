package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

const frameCostUSD = 0.001

// UsageMetricsHandler returns usage totals for the last 30 days.
func UsageMetricsHandler() gin.HandlerFunc {
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

		since := time.Now().UTC().AddDate(0, 0, -30)
		var totalFrames int64
		if err := gormDB.Model(&db.TelemetryUsage{}).
			Select("COALESCE(SUM(frames_processed), 0)").
			Where("tenant_id = ? AND timestamp >= ?", tenantID, since).
			Scan(&totalFrames).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load usage"})
			return
		}

		bill := float64(totalFrames) * frameCostUSD
		response := renderUsageCard(totalFrames, bill)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(response))
	}
}

func renderUsageCard(totalFrames int64, bill float64) string {
	escapedFrames := template.HTMLEscapeString(fmt.Sprintf("%d", totalFrames))
	escapedBill := template.HTMLEscapeString(fmt.Sprintf("$%.2f", bill))
	return "<div class=\"flex items-center justify-between\">" +
		"<div>" +
		"<h2 class=\"text-lg font-semibold\">Usage (Last 30 Days)</h2>" +
		"<p class=\"mt-1 text-sm text-slate-600\">Billing usage updated every 30 seconds.</p>" +
		"</div>" +
		"<div class=\"text-right\">" +
		"<div class=\"text-xs uppercase text-slate-500\">Frames Processed</div>" +
		"<div class=\"text-2xl font-semibold\">" + escapedFrames + "</div>" +
		"<div class=\"mt-2 text-xs uppercase text-slate-500\">Estimated Bill</div>" +
		"<div class=\"text-lg font-semibold text-slate-900\">" + escapedBill + "</div>" +
		"</div>" +
		"</div>"
}
