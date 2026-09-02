package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"lifeline_backend/internal/models"

	"github.com/gorilla/websocket"
)

// EphemeralMessage represents an in-memory chat message for an active request session
type EphemeralMessage struct {
	SenderID    uint      `json:"sender_id"`
	RecipientID uint      `json:"recipient_id"`
	RequestID   uint      `json:"request_id"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

// Client represents a connected pharmacy or patient WebSocket session
type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub tracks and manages active WebSocket connections and in-memory chat buffers
type Hub struct {
	clients      map[uint]*Client
	chatMessages map[uint][]EphemeralMessage // Key: request_id, Value: ephemeral message history
	Register     chan *Client
	Unregister   chan *Client
	mu           sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:      make(map[uint]*Client),
		chatMessages: make(map[uint][]EphemeralMessage),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
	}
}

// StoreChatMessage stores a message in the in-memory request buffer (max 100 per request)
func (h *Hub) StoreChatMessage(msg EphemeralMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	history := h.chatMessages[msg.RequestID]
	if len(history) >= 100 {
		history = history[1:]
	}
	h.chatMessages[msg.RequestID] = append(history, msg)
}

// GetChatMessages retrieves the in-memory messages for an active request
func (h *Hub) GetChatMessages(requestID uint) []EphemeralMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()

	history, exists := h.chatMessages[requestID]
	if !exists {
		return []EphemeralMessage{}
	}
	result := make([]EphemeralMessage, len(history))
	copy(result, history)
	return result
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			// Close old connection if exists
			if old, ok := h.clients[client.UserID]; ok {
				old.Conn.Close()
				close(old.Send)
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastToNearby sends the request payload to connected pharmacies that are in the matching subset
func (h *Hub) BroadcastToNearby(request models.EmergencyRequest, nearbyPharmacies []models.PharmacyProfile) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Map of pharmacy IDs that should receive the broadcast
	nearbyMap := make(map[uint]bool)
	for _, p := range nearbyPharmacies {
		nearbyMap[p.UserID] = true
	}

	payload, err := json.Marshal(map[string]interface{}{
		"event":   "NEW_EMERGENCY_REQUEST",
		"request": request,
	})
	if err != nil {
		return
	}

	for userID, client := range h.clients {
		if nearbyMap[userID] {
			select {
			case client.Send <- payload:
			default:
				// If send channel is blocked, clean up the client connection
				go func(c *Client) {
					c.Conn.Close()
				}(client)
			}
		}
	}
}

// SendToUser sends a direct message to a specific user (e.g. sending a response back to the patient)
func (h *Hub) SendToUser(userID uint, event string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, exists := h.clients[userID]
	if !exists {
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"event": event,
		"data":  data,
	})
	if err != nil {
		return
	}

	select {
	case client.Send <- payload:
	default:
		go func(c *Client) {
			c.Conn.Close()
		}(client)
	}
}
