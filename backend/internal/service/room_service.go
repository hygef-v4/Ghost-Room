package service

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"ghostroom/backend/internal/model"
)

type RoomService struct {
	rooms map[string]*model.Room
	mu    sync.RWMutex
}

func NewRoomService() *RoomService {
	return &RoomService{
		rooms: make(map[string]*model.Room),
	}
}

func (s *RoomService) CreateRoom(ttlMinutes int) (*model.Room, error) {
	if ttlMinutes <= 0 {
		ttlMinutes = 60 // Default 1 hour
	}

	roomID := uuid.New().String()
	room := model.NewRoom(roomID, ttlMinutes)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[roomID] = room

	return room, nil
}

func (s *RoomService) GetRoom(id string) (*model.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, exists := s.rooms[id]
	if !exists {
		return nil, errors.New("room not found")
	}

	if room.IsExpired() {
		// Should we delete it here?
		return nil, errors.New("room expired")
	}

	return room, nil
}

func (s *RoomService) DeleteRoom(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[id]; !exists {
		return errors.New("room not found")
	}

	delete(s.rooms, id)
	return nil
}

func (s *RoomService) CleanupExpiredRooms() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, room := range s.rooms {
		if room.IsExpired() {
			delete(s.rooms, id)
			count++
		}
	}
	return count
}
