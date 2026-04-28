package handlers

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/middleware"
	"github.com/sixpatrol/sixpatrol-server/queue"
)

// ProxyVideoHandler accepts a video chunk and publishes it for async processing.
func ProxyVideoHandler() gin.HandlerFunc {
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

		timestampStr := firstFormValue(c, "timestamp", "Timestamp")
		streamID := firstFormValue(c, "stream_id", "Stream_ID", "streamId")
		if timestampStr == "" || streamID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing timestamp or stream_id"})
			return
		}

		timestampValue, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp"})
			return
		}

		fileHeader, err := getUploadFile(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing video chunk"})
			return
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !isAllowedVideoExt(ext) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unsupported video format"})
			return
		}

		stored, err := storeProxyVideo(c.Request.Context(), fileHeader, streamID, ext)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to store video chunk"})
			return
		}

		msg := queue.ProxyVideoMessage{
			TenantID:       tenantID,
			FilePath:       stored.LocalPath,
			StorageBackend: stored.Backend,
			Bucket:         stored.Bucket,
			ObjectKey:      stored.ObjectKey,
			StreamID:       streamID,
			Timestamp:      time.Unix(timestampValue, 0).UTC(),
		}
		if err := queue.PublishProxyVideo(msg); err != nil {
			_ = deleteStoredProxyVideo(c.Request.Context(), stored)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "queue unavailable"})
			return
		}

		c.Status(http.StatusAccepted)
	}
}

func firstFormValue(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.PostForm(key)); value != "" {
			return value
		}
	}
	return ""
}

func getUploadFile(c *gin.Context) (*multipart.FileHeader, error) {
	fileHeader, err := c.FormFile("chunk")
	if err == nil {
		return fileHeader, nil
	}
	return c.FormFile("file")
}

func isAllowedVideoExt(ext string) bool {
	switch ext {
	case ".mp4", ".webm":
		return true
	default:
		return false
	}
}

func sanitizeSegment(value string) string {
	if value == "" {
		return "stream"
	}

	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}

	sanitized := builder.String()
	if len(sanitized) > 48 {
		sanitized = sanitized[:48]
	}
	return sanitized
}

func saveUploadedFile(fileHeader *multipart.FileHeader, destPath string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	output, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, file)
	return err
}
