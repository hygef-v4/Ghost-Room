package ws

import (
	"log"
	"sync"
)

// Hub manages all websocket clients grouped by room
type Hub struct {
	// rooms maps roomID -> { clientID -> *WSClient }
	rooms map[string]map[string]*WSClient
	mu    sync.RWMutex

	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan *RoomMessage
}

type RoomMessage struct {
	RoomID  string
	Payload []byte
	From    string
	To      string // empty = broadcast to all in room
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[string]*WSClient),
		register:   make(chan *WSClient, 64),
		unregister: make(chan *WSClient, 64),
		broadcast:  make(chan *RoomMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.RoomID] == nil {
				h.rooms[client.RoomID] = make(map[string]*WSClient)
			}
			h.rooms[client.RoomID][client.ID] = client
			roomSize := len(h.rooms[client.RoomID])
			h.mu.Unlock()
			log.Printf("[Hub] Client %s joined room %s (total: %d)", client.ID, client.RoomID, roomSize)

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.RoomID]; ok {
				delete(room, client.ID)
				if len(room) == 0 {
					delete(h.rooms, client.RoomID)
				}
			}
			h.mu.Unlock()
			close(client.send)
			log.Printf("[Hub] Client %s left room %s", client.ID, client.RoomID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			room, ok := h.rooms[msg.RoomID]
			h.mu.RUnlock()
			if !ok {
				continue
			}

			if msg.To != "" {
				// Directed message – send only to target
				if target, ok := room[msg.To]; ok {
					select {
					case target.send <- msg.Payload:
					default:
						log.Printf("[Hub] Dropped message to %s (buffer full)", msg.To)
					}
				}
			} else {
				// Broadcast to everyone except sender
				for id, client := range room {
					if id == msg.From {
						continue
					}
					select {
					case client.send <- msg.Payload:
					default:
						log.Printf("[Hub] Dropped broadcast to %s (buffer full)", id)
					}
				}
			}
		}
	}
}

// Helpers for the handler

func (h *Hub) Register(client *WSClient) {
	h.register <- client
}

func (h *Hub) Unregister(client *WSClient) {
	h.unregister <- client
}

func (h *Hub) Broadcast(msg *RoomMessage) {
	h.broadcast <- msg
}

func (h *Hub) RoomClients(roomID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := []string{}
	for id := range h.rooms[roomID] {
		ids = append(ids, id)
	}
	return ids
}
