package dto

type CreateRoomRequest struct {
	TTLMinutes int `json:"ttl_minutes,omitempty"`
	MaxUsers   int `json:"max_users,omitempty"`
}

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
	Link   string `json:"link"`
}

type JoinRoomRequest struct {
	RoomID string `json:"room_id"`
}

type JoinRoomResponse struct {
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
	Clients int `json:"clients"`
}

type WebSocketMessage struct {
	Type    string          `json:"type"` // offer, answer, ice, join, leave, chat
	RoomID  string          `json:"room_id"`
	From    string          `json:"from"`
	To      string          `json:"to,omitempty"`
	Payload interface{}     `json:"payload,omitempty"`
}

type ICECandidate struct {
	Candidate        string `json:"candidate"`
	SdpMLineIndex    *int   `json:"sdp_mline_index"`
	SdpMid           string `json:"sdp_mid,omitempty"`
}

type OfferAnswerPayload struct {
	Type string `json:"type"` // offer or answer
	SDP  string `json:"sdp"`
}

type ChatMessage struct {
	Message string `json:"message"`
	Sender  string `json:"sender"`
}
