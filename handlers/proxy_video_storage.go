package handlers

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/env"
)

const (
	proxyVideoStorageLocal = "local"
	proxyVideoStorageS3    = "s3"
)

type storedProxyVideo struct {
	Backend   string
	LocalPath string
	Bucket    string
	ObjectKey string
}

func storeProxyVideo(ctx context.Context, fileHeader *multipart.FileHeader, streamID, ext string) (storedProxyVideo, error) {
	backend := strings.ToLower(strings.TrimSpace(env.Get("PROXY_VIDEO_STORAGE", "")))
	if backend == "" {
		if strings.TrimSpace(env.Get("PROXY_VIDEO_BUCKET", "")) != "" {
			backend = proxyVideoStorageS3
		} else {
			backend = proxyVideoStorageLocal
		}
	}
	switch backend {
	case "", proxyVideoStorageLocal:
		return storeProxyVideoLocal(fileHeader, streamID, ext)
	case proxyVideoStorageS3, "r2":
		return storeProxyVideoS3(ctx, fileHeader, streamID, ext)
	default:
		return storedProxyVideo{}, fmt.Errorf("unknown proxy video storage backend %q", backend)
	}
}

func storeProxyVideoLocal(fileHeader *multipart.FileHeader, streamID, ext string) (storedProxyVideo, error) {
	storageDir := env.Get("PROXY_VIDEO_DIR", "tmp/proxy-video")
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return storedProxyVideo{}, err
	}

	objectName := buildProxyVideoObjectName(streamID, ext)
	destPath := filepath.Join(storageDir, objectName)
	if err := saveUploadedFile(fileHeader, destPath); err != nil {
		return storedProxyVideo{}, err
	}

	return storedProxyVideo{
		Backend:   proxyVideoStorageLocal,
		LocalPath: destPath,
	}, nil
}

func storeProxyVideoS3(ctx context.Context, fileHeader *multipart.FileHeader, streamID, ext string) (storedProxyVideo, error) {
	bucket := strings.TrimSpace(env.Get("PROXY_VIDEO_BUCKET", ""))
	if bucket == "" {
		return storedProxyVideo{}, fmt.Errorf("PROXY_VIDEO_BUCKET is required for s3 storage")
	}

	client, err := newS3Client(ctx)
	if err != nil {
		return storedProxyVideo{}, err
	}

	objectKey := buildProxyVideoObjectName(streamID, ext)
	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	file, err := fileHeader.Open()
	if err != nil {
		return storedProxyVideo{}, err
	}
	defer file.Close()

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Body:   file,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if fileHeader.Size > 0 {
		size := fileHeader.Size
		input.ContentLength = &size
	}

	if _, err := client.PutObject(ctx, input); err != nil {
		return storedProxyVideo{}, err
	}

	return storedProxyVideo{
		Backend:   proxyVideoStorageS3,
		Bucket:    bucket,
		ObjectKey: objectKey,
	}, nil
}

func deleteStoredProxyVideo(ctx context.Context, stored storedProxyVideo) error {
	switch stored.Backend {
	case "", proxyVideoStorageLocal:
		if stored.LocalPath == "" {
			return nil
		}
		if err := os.Remove(stored.LocalPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	case proxyVideoStorageS3:
		if stored.Bucket == "" || stored.ObjectKey == "" {
			return nil
		}
		client, err := newS3Client(ctx)
		if err != nil {
			return err
		}
		_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(stored.Bucket),
			Key:    aws.String(stored.ObjectKey),
		})
		return err
	default:
		return fmt.Errorf("unknown proxy video storage backend %q", stored.Backend)
	}
}

func newS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := strings.TrimSpace(env.Get("AWS_ENDPOINT_URL", ""))
	if endpoint == "" {
		return s3.NewFromConfig(cfg), nil
	}

	return s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.EndpointResolver = s3.EndpointResolverFromURL(endpoint)
		options.UsePathStyle = true
	}), nil
}

func buildProxyVideoObjectName(streamID, ext string) string {
	safeStreamID := sanitizeSegment(streamID)
	return fmt.Sprintf("%s_%s%s", safeStreamID, uuid.New().String(), ext)
}
