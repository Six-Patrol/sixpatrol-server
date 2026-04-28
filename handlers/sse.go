package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/middleware"
	"github.com/sixpatrol/sixpatrol-server/realtime"
)

// PiracyStreamHandler streams SSE updates for piracy detections.
func PiracyStreamHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "stream unsupported"})
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		messages, unsubscribe := realtime.Subscribe(tenantID)
		defer unsubscribe()

		notify := c.Request.Context().Done()
		for {
			select {
			case <-notify:
				return
			case msg, ok := <-messages:
				if !ok {
					return
				}
				writeSSE(c.Writer, "message", msg)
				flusher.Flush()
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, event string, data string) {
	if event != "" {
		_, _ = fmt.Fprintf(w, "event: %s\n", event)
	}
	for _, line := range strings.Split(data, "\n") {
		_, _ = fmt.Fprintf(w, "data: %s\n", line)
	}
	_, _ = fmt.Fprint(w, "\n")
}
