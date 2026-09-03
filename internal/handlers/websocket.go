package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"lifeline_backend/internal/utils"
	wshub "lifeline_backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	Hub       *wshub.Hub
	JWTSecret string
}

func NewWSHandler(hub *wshub.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{
		Hub:       hub,
		JWTSecret: jwtSecret,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for MVP
	},
}

// ConnectPharmacy upgrades the HTTP connection to a WebSocket for users (pharmacies or patients)
func (h *WSHandler) ConnectPharmacy(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token query parameter is required"})
		return
	}

	claims, err := utils.ParseToken(tokenString, h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	userIDFloat, ok1 := (*claims)["user_id"].(float64)
	role, ok2 := (*claims)["role"].(string)
	if !ok1 || !ok2 || (role != "PHARMACY" && role != "PATIENT") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Unauthorized connection"})
		return
	}

	userID := uint(userIDFloat)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// HTTP upgrade error is handled internally by Upgrade
		return
	}

	client := &wshub.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h.Hub.Register <- client

	// Start read & write loops in goroutines
	go h.writePump(client)
	go h.readPump(client)
}

// GetChatMessages retrieves buffered in-memory messages for an active request
func (h *WSHandler) GetChatMessages(c *gin.Context) {
	reqIDStr := c.Param("id")
	reqID, err := strconv.ParseUint(reqIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	messages := h.Hub.GetChatMessages(uint(reqID))
	c.JSON(http.StatusOK, messages)
}

func (h *WSHandler) writePump(client *wshub.Client) {
	defer func() {
		client.Conn.Close()
	}()

	for {
		message, ok := <-client.Send
		if !ok {
			// Hub closed the channel
			_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

func (h *WSHandler) readPump(client *wshub.Client) {
	defer func() {
		h.Hub.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse the incoming message
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		event, _ := msg["event"].(string)
		if event == "CHAT_SEND" {
			recipientIDFloat, _ := msg["recipient_id"].(float64)
			requestIDFloat, _ := msg["request_id"].(float64)
			content, _ := msg["content"].(string)

			recipientID := uint(recipientIDFloat)
			requestID := uint(requestIDFloat)

			if recipientID > 0 && requestID > 0 && content != "" {
				now := time.Now()
				// Store in in-memory session buffer
				h.Hub.StoreChatMessage(wshub.EphemeralMessage{
					SenderID:    client.UserID,
					RecipientID: recipientID,
					RequestID:   requestID,
					Content:     content,
					CreatedAt:   now,
				})

				relayData := map[string]interface{}{
					"sender_id":  client.UserID,
					"request_id": requestID,
					"content":    content,
					"created_at": now,
				}
				h.Hub.SendToUser(recipientID, "CHAT_RECEIVE", relayData)
			}
		}
	}
}
