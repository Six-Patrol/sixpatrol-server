package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sixpatrol/sixpatrol-server/db"
	"github.com/sixpatrol/sixpatrol-server/env"
	"github.com/sixpatrol/sixpatrol-server/handlers"
	"github.com/sixpatrol/sixpatrol-server/middleware"
)

func main() {
	// Load .env if present (development convenience).
	_ = env.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Initialize CockroachDB connection (reads COCKROACH_DSN).
	gormDB, err := db.InitCockroachDB(ctx)
	if err != nil {
		log.Fatalf("failed to initialize CockroachDB: %v", err)
	}
	db.SetDB(gormDB)

	// Run auto-migrations to ensure the schema exists.
	if err := db.AutoMigrate(gormDB); err != nil {
		log.Fatalf("failed to auto-migrate database schema: %v", err)
	}

	// Initialize Qdrant client (reads QDRANT_HOST/PORT and optional QDRANT_API_KEY).
	qdrantClient, err := db.InitQdrantClient()
	if err != nil {
		log.Fatalf("failed to initialize Qdrant client: %v", err)
	}
	_ = qdrantClient

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/static", "./static")

	router.GET("/dashboard", handlers.DashboardHandler())

	ingestGroup := router.Group("/ingest/v1")
	ingestGroup.Use(middleware.HMACAuthMiddleware())
	ingestGroup.POST("/telemetry", handlers.TelemetryIngestHandler())
	ingestGroup.POST("/proxy-video", handlers.ProxyVideoHandler())

	apiGroup := router.Group("/api/v1")
	apiGroup.Use(middleware.JWTAuthMiddleware())
	apiGroup.POST("/keys/generate", handlers.GenerateAPIKeyHandler())
	apiGroup.GET("/metrics/usage", handlers.UsageMetricsHandler())
	apiGroup.GET("/stream/piracy", handlers.PiracyStreamHandler())

	internalGroup := router.Group("/api/v1/internal")
	internalGroup.POST("/match-found", handlers.InternalMatchFoundHandler())

	port := env.Get("PORT", "8080")
	log.Printf("Gin server listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
