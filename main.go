package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nikhilsahni7/fast-chat/auth"
	"github.com/nikhilsahni7/fast-chat/chat"
	"github.com/nikhilsahni7/fast-chat/database"
	"github.com/nikhilsahni7/fast-chat/user"
	"github.com/rs/cors"
)

func main() {
	database.InitDB()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Setup CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	r.Use(corsMiddleware.Handler)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)
	})

	r.Route("/users", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/", user.ListUsers)
		r.Get("/{userID}", user.GetUser)
		r.Put("/{userID}", user.UpdateUser)
		r.Delete("/{userID}", user.DeleteUser)
	})

	r.Get("/ws", chat.HandleWebSocket)
	// r.Post("/chat/message", chat.SendMessage)

	r.Route("/chat", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		r.Get("/history/{receiverID}", chat.GetChatHistory)
		r.Get("/conversations", chat.GetConversations)
		r.Post("/message", chat.SendMessage)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
