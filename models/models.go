package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string    `json:"username" gorm:"uniqueIndex"`
	Email        string    `json:"email" gorm:"uniqueIndex"`
	Password     string    `json:"-"`
	LastSeen     time.Time `json:"last_seen"`
	IsOnline     bool      `json:"is_online"`
	ProfileImage string    `json:"profile_image"`
}

type Message struct {
	gorm.Model
	SenderID   uint       `json:"sender_id"`
	ReceiverID uint       `json:"receiver_id"`
	Content    string     `json:"content"`
	Timestamp  time.Time  `json:"timestamp"`
	ReadAt     *time.Time `json:"read_at"`
	Type       string     `json:"type"` // text, image, file, etc.
}

type Conversation struct {
	gorm.Model
	User1ID       uint     `json:"user1_id"`
	User2ID       uint     `json:"user2_id"`
	LastMessageID uint     `json:"last_message_id"`
	UnreadCount   int      `json:"unread_count"`
	LastMessage   *Message `json:"last_message" gorm:"foreignKey:LastMessageID"`
}

type TypingStatus struct {
	UserID     uint      `json:"user_id"`
	ReceiverID uint      `json:"receiver_id"`
	IsTyping   bool      `json:"is_typing"`
	Timestamp  time.Time `json:"timestamp"`
}
