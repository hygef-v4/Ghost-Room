package model

import (
	"sync"
	"time"
)

type Room struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
	Status    string
	MaxUsers  int
	Clients   map[string]*Client
	mu        sync.RWMutex
}

type Client struct {
	ID     string
	RoomID string
	UserID string
	Conn   interface{}
}

func NewRoom(id string, ttlMinutes int) *Room {
	return &Room{
		ID:        id,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(ttlMinutes) * time.Minute),
		Status:    "active",
		MaxUsers:  10,
		Clients:   make(map[string]*Client),
	}
}

func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[client.ID] = client
}

func (r *Room) RemoveClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, clientID)
}

func (r *Room) GetClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clients := make([]*Client, 0, len(r.Clients))
	for _, client := range r.Clients {
		clients = append(clients, client)
	}
	return clients
}

func (r *Room) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func (r *Room) IsFull() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Clients) >= r.MaxUsers
}
