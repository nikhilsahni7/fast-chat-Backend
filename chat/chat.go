package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/nikhilsahni7/fast-chat/auth"
	"github.com/nikhilsahni7/fast-chat/database"
	"github.com/nikhilsahni7/fast-chat/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // not recommended for production we will use correct otigin check
	},
}

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	userID uint
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

var hub = Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			h.broadcastUserStatus(client.userID, true)
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.broadcastUserStatus(client.userID, false)
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) broadcastUserStatus(userID uint, isOnline bool) {
	status := WebSocketMessage{
		Type: "user_status",
		Payload: map[string]interface{}{
			"user_id":   userID,
			"is_online": isOnline,
		},
	}
	statusJSON, _ := json.Marshal(status)
	h.broadcast <- statusJSON
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{conn: conn, send: make(chan []byte, 256), userID: userID}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
		updateUserStatus(c.userID, false)
	}()

	c.conn.SetReadLimit(1024)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("error decoding message: %v", err)
			continue
		}

		switch wsMessage.Type {
		case "chat_message":
			handleChatMessage(c, wsMessage.Payload)
		case "typing_status":
			handleTypingStatus(c, wsMessage.Payload)
		}
	}
}
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func handleChatMessage(c *Client, payload interface{}) {
	var message models.Message
	payloadJSON, _ := json.Marshal(payload)
	if err := json.Unmarshal(payloadJSON, &message); err != nil {
		log.Printf("error decoding chat message: %v", err)
		return
	}

	message.SenderID = c.userID
	message.Timestamp = time.Now()

	db := database.GetDB()
	if err := db.Create(&message).Error; err != nil {
		log.Printf("error saving message: %v", err)
		return
	}

	updateConversation(message.SenderID, message.ReceiverID, message.ID)

	messageJSON, _ := json.Marshal(WebSocketMessage{
		Type:    "chat_message",
		Payload: message,
	})

	// Send to all connected clients
	hub.broadcast <- messageJSON
}
func handleTypingStatus(c *Client, payload interface{}) {
	var typingStatus struct {
		ReceiverID uint `json:"receiver_id"`
		IsTyping   bool `json:"is_typing"`
	}
	payloadJSON, _ := json.Marshal(payload)
	if err := json.Unmarshal(payloadJSON, &typingStatus); err != nil {
		log.Printf("error decoding typing status: %v", err)
		return
	}

	statusJSON, _ := json.Marshal(WebSocketMessage{
		Type: "typing_status",
		Payload: map[string]interface{}{
			"sender_id":   c.userID,
			"receiver_id": typingStatus.ReceiverID,
			"is_typing":   typingStatus.IsTyping,
		},
	})
	hub.broadcast <- statusJSON
}

func updateConversation(user1ID, user2ID, lastMessageID uint) {
	db := database.GetDB()
	var conversation models.Conversation
	result := db.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
		user1ID, user2ID, user2ID, user1ID).First(&conversation)

	if result.Error != nil {
		conversation = models.Conversation{
			User1ID:       user1ID,
			User2ID:       user2ID,
			LastMessageID: lastMessageID,
			UnreadCount:   1,
		}
		db.Create(&conversation)
	} else {
		conversation.LastMessageID = lastMessageID
		conversation.UnreadCount++
		db.Save(&conversation)
	}
}

func updateUserStatus(userID uint, isOnline bool) {
	db := database.GetDB()
	db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_online": isOnline,
		"last_seen": time.Now(),
	})
}

func GetChatHistory(w http.ResponseWriter, r *http.Request) {
	senderID := uint(r.Context().Value(auth.UserID(0)).(auth.UserID))
	receiverID, _ := strconv.ParseUint(chi.URLParam(r, "receiverID"), 10, 32)

	db := database.GetDB()
	var messages []models.Message
	if err := db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		senderID, receiverID, receiverID, senderID).
		Order("created_at asc").
		Find(&messages).Error; err != nil {
		http.Error(w, "Error fetching chat history", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	db.Model(&models.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND read_at IS NULL", receiverID, senderID).
		Update("read_at", now)

	db.Model(&models.Conversation{}).
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
			senderID, receiverID, receiverID, senderID).
		Update("unread_count", 0)

	json.NewEncoder(w).Encode(messages)
}

func GetConversations(w http.ResponseWriter, r *http.Request) {
	userID := uint(r.Context().Value(auth.UserID(0)).(auth.UserID))

	db := database.GetDB()
	var conversations []models.Conversation
	if err := db.Where("user1_id = ? OR user2_id = ?", userID, userID).
		Preload("LastMessage").
		Find(&conversations).Error; err != nil {
		http.Error(w, "Error fetching conversations", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(conversations)
}

func SendMessage(w http.ResponseWriter, r *http.Request) {
	var message models.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := uint(r.Context().Value(auth.UserID(0)).(auth.UserID))
	message.SenderID = userID
	message.Timestamp = time.Now()

	db := database.GetDB()
	if err := db.Create(&message).Error; err != nil {
		http.Error(w, "Error sending message", http.StatusInternalServerError)
		return
	}

	updateConversation(message.SenderID, message.ReceiverID, message.ID)

	messageJSON, _ := json.Marshal(WebSocketMessage{
		Type:    "chat_message",
		Payload: message,
	})
	hub.broadcast <- messageJSON

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func init() {
	go hub.run()
	fmt.Println("Chat hub running")
}
