package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"ghostroom/backend/api"
	"ghostroom/backend/internal/config"
	"ghostroom/backend/internal/service"
	"ghostroom/backend/internal/ws"
)

func main() {
	// Load config
	cfg := config.Load()

	// Setup Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.AllowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize services
	roomService := service.NewRoomService()
	hub := ws.NewHub()

	// Start hub
	go hub.Run()

	// Start cleanup goroutine
	go cleanupWorker(roomService)

	// Setup routes
	baseURL := fmt.Sprintf("http://localhost:%s", cfg.Port)
	api.SetupRoutes(router, roomService, hub, baseURL)

	// Start server
	log.Printf("Starting server on port %s (env: %s)", cfg.Port, cfg.Environment)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func cleanupWorker(roomService *service.RoomService) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		count := roomService.CleanupExpiredRooms()
		if count > 0 {
			log.Printf("[Cleanup] Removed %d expired rooms", count)
		}
	}
}
