package database

import (
	"log"
	"sync"

	"github.com/nikhilsahni7/fast-chat/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func InitDB() {
	once.Do(func() {
		dsn := "host=localhost user=postgres password=postgres dbname=fastchat port=5432 sslmode=disable TimeZone=Asia/Kolkata" // or use yout own dsn
		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Auto-migrate the schema
		err = db.AutoMigrate(&models.User{}, &models.Message{}, &models.Conversation{}, &models.TypingStatus{})
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		log.Println("Database migrated successfully")
		log.Println("Connected to database")
	})
}

func GetDB() *gorm.DB {
	if db == nil {
		InitDB()
	}
	return db
}
