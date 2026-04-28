package tests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/handlers"
	"github.com/sixpatrol/sixpatrol-server/middleware"
	"github.com/sixpatrol/sixpatrol-server/queue"
)

func TestProxyVideoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant", func(t *testing.T) {
		router := gin.New()
		router.POST("/proxy", handlers.ProxyVideoHandler())

		body, contentType := newMultipartRequest(t, map[string]string{"timestamp": "123", "stream_id": "stream-a"}, "chunk", "clip.mp4", "data")
		req := httptest.NewRequest(http.MethodPost, "/proxy", body)
		req.Header.Set("Content-Type", contentType)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		router := gin.New()
		router.POST("/proxy", func(c *gin.Context) {
			c.Set(middleware.TenantIDContextKey, uuid.New())
			handlers.ProxyVideoHandler()(c)
		})

		body, contentType := newMultipartRequest(t, map[string]string{"timestamp": "123", "stream_id": "stream-a"}, "", "", "")
		req := httptest.NewRequest(http.MethodPost, "/proxy", body)
		req.Header.Set("Content-Type", contentType)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.Code)
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		router := gin.New()
		router.POST("/proxy", func(c *gin.Context) {
			c.Set(middleware.TenantIDContextKey, uuid.New())
			handlers.ProxyVideoHandler()(c)
		})

		body, contentType := newMultipartRequest(t, map[string]string{"timestamp": "123", "stream_id": "stream-a"}, "chunk", "clip.txt", "data")
		req := httptest.NewRequest(http.MethodPost, "/proxy", body)
		req.Header.Set("Content-Type", contentType)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.Code)
		}
	})

	t.Run("missing metadata", func(t *testing.T) {
		router := gin.New()
		router.POST("/proxy", func(c *gin.Context) {
			c.Set(middleware.TenantIDContextKey, uuid.New())
			handlers.ProxyVideoHandler()(c)
		})

		body, contentType := newMultipartRequest(t, map[string]string{"stream_id": "stream-a"}, "chunk", "clip.mp4", "data")
		req := httptest.NewRequest(http.MethodPost, "/proxy", body)
		req.Header.Set("Content-Type", contentType)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.Code)
		}
	})

	t.Run("queue full", func(t *testing.T) {
		queueChan := make(chan queue.ProxyVideoMessage, 1)
		queueChan <- queue.ProxyVideoMessage{}
		previous := queue.SetProxyVideoQueue(queueChan)
		defer queue.SetProxyVideoQueue(previous)

		tempDir := t.TempDir()
		t.Setenv("PROXY_VIDEO_DIR", tempDir)

		router := gin.New()
		router.POST("/proxy", func(c *gin.Context) {
			c.Set(middleware.TenantIDContextKey, uuid.New())
			handlers.ProxyVideoHandler()(c)
		})

		body, contentType := newMultipartRequest(t, map[string]string{"timestamp": "123", "stream_id": "stream-a"}, "chunk", "clip.mp4", "data")
		req := httptest.NewRequest(http.MethodPost, "/proxy", body)
		req.Header.Set("Content-Type", contentType)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", resp.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		queueChan := make(chan queue.ProxyVideoMessage, 1)
		previous := queue.SetProxyVideoQueue(queueChan)
		defer queue.SetProxyVideoQueue(previous)

		tempDir := t.TempDir()
		t.Setenv("PROXY_VIDEO_DIR", tempDir)

		tenantID := uuid.New()
		router := gin.New()
		router.POST("/proxy", func(c *gin.Context) {
			c.Set(middleware.TenantIDContextKey, tenantID)
			handlers.ProxyVideoHandler()(c)
		})

		body, contentType := newMultipartRequest(t, map[string]string{"timestamp": "123", "stream_id": "stream-a"}, "chunk", "clip.webm", "data")
		req := httptest.NewRequest(http.MethodPost, "/proxy", body)
		req.Header.Set("Content-Type", contentType)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d", resp.Code)
		}

		select {
		case msg := <-queueChan:
			if msg.TenantID != tenantID {
				t.Fatalf("expected tenant ID %s, got %s", tenantID, msg.TenantID)
			}
			if msg.StreamID != "stream-a" {
				t.Fatalf("expected stream_id stream-a, got %s", msg.StreamID)
			}
			if !strings.HasSuffix(msg.FilePath, ".webm") {
				t.Fatalf("expected .webm file, got %s", msg.FilePath)
			}
			if _, err := os.Stat(msg.FilePath); err != nil {
				t.Fatalf("expected file to exist: %v", err)
			}
			_ = os.Remove(msg.FilePath)
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("expected message to be published")
		}
	})
}

func newMultipartRequest(t *testing.T, fields map[string]string, fileField, fileName, fileData string) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("failed to write field: %v", err)
		}
	}

	if fileField != "" {
		part, err := writer.CreateFormFile(fileField, fileName)
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		if _, err := part.Write([]byte(fileData)); err != nil {
			t.Fatalf("failed to write form data: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	return &body, writer.FormDataContentType()
}
