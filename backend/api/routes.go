package api

import (
	"github.com/gin-gonic/gin"
	httpHandler "ghostroom/backend/internal/handler/http"
	wsHandler "ghostroom/backend/internal/handler/ws"
	"ghostroom/backend/internal/service"
	"ghostroom/backend/internal/ws"
)

func SetupRoutes(router *gin.Engine, roomService *service.RoomService, hub *ws.Hub, baseURL string) {
	// HTTP handlers
	roomHandler := httpHandler.NewRoomHandler(roomService, baseURL)
	wsHandlerInstance := wsHandler.NewWSHandler(hub, roomService)

	// API routes
	api := router.Group("/api")
	{
		api.POST("/rooms", roomHandler.CreateRoom)
		api.GET("/rooms/:id", roomHandler.GetRoom)
		api.POST("/rooms/:id/join", roomHandler.JoinRoom)
	}

	// WebSocket route
	router.GET("/ws", wsHandlerInstance.HandleWS)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
