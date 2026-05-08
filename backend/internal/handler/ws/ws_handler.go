package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"ghostroom/backend/internal/dto"
	"ghostroom/backend/internal/service"
	domainws "ghostroom/backend/internal/ws"
)

type WSHandler struct {
	hub         *domainws.Hub
	roomService *service.RoomService
	upgrader    websocket.Upgrader
}

func NewWSHandler(hub *domainws.Hub, roomService *service.RoomService) *WSHandler {
	return &WSHandler{
		hub:         hub,
		roomService: roomService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, validate origin properly
				return true
			},
		},
	}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	roomID := c.Query("roomId")
	userID := c.Query("userId")

	if roomID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId and userId required"})
		return
	}

	// Verify room exists
	room, err := h.roomService.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	if room.IsFull() {
		c.JSON(http.StatusConflict, gin.H{"error": "room is full"})
		return
	}

	// Upgrade connection
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WSHandler] Upgrade error: %v", err)
		return
	}

	clientID := uuid.NewString()
	client := domainws.NewWSClient(clientID, roomID, userID, conn, h.hub)
	h.hub.Register(client)

	// Send join notification to others
	joinMsg := dto.WebSocketMessage{
		Type:   "join",
		RoomID: roomID,
		From:   clientID,
		Payload: map[string]interface{}{
			"user_id": userID,
			"clients": h.hub.RoomClients(roomID),
		},
	}
	joinData, _ := json.Marshal(joinMsg)
	h.hub.Broadcast(&domainws.RoomMessage{
		RoomID:  roomID,
		Payload: joinData,
		From:    clientID,
	})

	// Start read/write pumps
	go client.WritePump()
	client.ReadPump(h.handleMessage)
}

func (h *WSHandler) handleMessage(client *domainws.WSClient, data []byte) {
	var msg dto.WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[WSHandler] JSON unmarshal error: %v", err)
		return
	}

	msg.RoomID = client.RoomID
	msg.From = client.ID

	switch msg.Type {
	case "offer", "answer":
		// Relay to target peer
		if msg.To != "" {
			msgData, _ := json.Marshal(msg)
			h.hub.Broadcast(&domainws.RoomMessage{
				RoomID:  client.RoomID,
				Payload: msgData,
				From:    client.ID,
				To:      msg.To,
			})
		}

	case "ice":
		// Relay ICE candidate to target
		if msg.To != "" {
			msgData, _ := json.Marshal(msg)
			h.hub.Broadcast(&domainws.RoomMessage{
				RoomID:  client.RoomID,
				Payload: msgData,
				From:    client.ID,
				To:      msg.To,
			})
		}

	case "chat":
		// Broadcast chat to all in room
		msgData, _ := json.Marshal(msg)
		h.hub.Broadcast(&domainws.RoomMessage{
			RoomID:  client.RoomID,
			Payload: msgData,
			From:    client.ID,
		})

	case "leave":
		// Client is leaving
		log.Printf("[WSHandler] Client %s leaving room %s", client.ID, client.RoomID)
		leaveMsg := dto.WebSocketMessage{
			Type:   "leave",
			RoomID: client.RoomID,
			From:   client.ID,
		}
		leaveData, _ := json.Marshal(leaveMsg)
		h.hub.Broadcast(&domainws.RoomMessage{
			RoomID:  client.RoomID,
			Payload: leaveData,
			From:    client.ID,
		})
		client.Close()

	default:
		log.Printf("[WSHandler] Unknown message type: %s", msg.Type)
	}
}
