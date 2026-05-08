package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ghostroom/backend/internal/dto"
	"ghostroom/backend/internal/service"
)

type RoomHandler struct {
	roomService *service.RoomService
	baseURL     string
}

func NewRoomHandler(roomService *service.RoomService, baseURL string) *RoomHandler {
	return &RoomHandler{roomService: roomService, baseURL: baseURL}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	room, err := h.roomService.CreateRoom(req.TTLMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room"})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateRoomResponse{
		RoomID: room.ID,
		Link:   h.baseURL + "/room/" + room.ID,
	})
}

func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("id")
	room, err := h.roomService.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         room.ID,
		"created_at": room.CreatedAt,
		"expires_at": room.ExpiresAt,
		"status":     room.Status,
		"max_users":  room.MaxUsers,
		"clients":    len(room.GetClients()),
	})
}

func (h *RoomHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	room, err := h.roomService.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if room.IsFull() {
		c.JSON(http.StatusConflict, gin.H{"error": "room is full"})
		return
	}

	c.JSON(http.StatusOK, dto.JoinRoomResponse{
		UserID:  uuid.NewString(),
		RoomID:  room.ID,
		Clients: len(room.GetClients()),
	})
}
