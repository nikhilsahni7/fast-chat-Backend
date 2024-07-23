# Fast-Chat

Fast-Chat is a real-time super fast chat application built using Go, Chi, PostgreSQL, and JWT for authentication. It supports real-time messaging, online status, typing status, and will include features for end-to-end encryption, file, and image uploads.

## Features

- **Real-time messaging** using WebSockets
- **User authentication** with JWT tokens
- **Online status tracking**
- **Typing status indication**
- **End-to-end encryption** (upcoming)
- **File and image uploads** (upcoming)

## Technologies Used

- **Backend:** Go, Chi
- **Database:** PostgreSQL
- **Authentication:** JWT
- **Real-time Communication:** WebSockets

## Models

### User

```go
type User struct {
	gorm.Model
	Username     string    `json:"username" gorm:"uniqueIndex"`
	Email        string    `json:"email" gorm:"uniqueIndex"`
	Password     string    `json:"-"`
	LastSeen     time.Time `json:"last_seen"`
	IsOnline     bool      `json:"is_online"`
	ProfileImage string    `json:"profile_image"`
}
```

### Message

```go


type Message struct {
	gorm.Model
	SenderID   uint       `json:"sender_id"`
	ReceiverID uint       `json:"receiver_id"`
	Content    string     `json:"content"`
	Timestamp  time.Time  `json:"timestamp"`
	ReadAt     *time.Time `json:"read_at"`
	Type       string     `json:"type"` // text, image, file, etc.
}
```

### Conversation

```go
type Conversation struct {
	gorm.Model
	User1ID       uint     `json:"user1_id"`
	User2ID       uint     `json:"user2_id"`
	LastMessageID uint     `json:"last_message_id"`
	UnreadCount   int      `json:"unread_count"`
	LastMessage   *Message `json:"last_message" gorm:"foreignKey:LastMessageID"`
}

```

### TypingStatus

```go
type TypingStatus struct {
	UserID     uint      `json:"user_id"`
	ReceiverID uint      `json:"receiver_id"`
	IsTyping   bool      `json:"is_typing"`
	Timestamp  time.Time `json:"timestamp"`
}

```

# Setup Instructions

### Clone the repository:

````
git clone https://github.com/nikhilsahni7/fast-chat-Backend.git

cd fast-chat

Install dependencies:
go mod tidy

Setup PostgreSQL database:
Create a new PostgreSQL database.
Update the database connection string in your configuration file.
Run the application:
go run main.go

Access the application:
The application will be running on http://localhost:8080.

Contributing
Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

License
This project is licensed under the MIT License. See the LICENSE file for details. ```











````
