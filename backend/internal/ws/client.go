package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 65536 // 64KB
)

type WSClient struct {
	ID     string
	RoomID string
	UserID string

	conn *websocket.Conn
	hub  *Hub
	send chan []byte
}

func NewWSClient(id, roomID, userID string, conn *websocket.Conn, hub *Hub) *WSClient {
	return &WSClient{
		ID:     id,
		RoomID: roomID,
		UserID: userID,
		conn:   conn,
		hub:    hub,
		send:   make(chan []byte, 256),
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *WSClient) ReadPump(onMessage func(client *WSClient, data []byte)) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WSClient] Unexpected close: %v", err)
			}
			break
		}
		onMessage(c, message)
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *WSClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain queued messages into the same writer
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Close gracefully closes the websocket connection
func (c *WSClient) Close() error {
	return c.conn.Close()
}
