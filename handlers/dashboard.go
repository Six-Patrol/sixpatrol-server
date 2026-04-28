package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DashboardHandler renders the dashboard HTML.
func DashboardHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "base", gin.H{
			"Title": "SixPatrol Dashboard",
		})
	}
}
