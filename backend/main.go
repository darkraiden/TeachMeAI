package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"teachme/internal/config"
	"teachme/internal/database"
	"teachme/internal/handlers"
	"teachme/internal/ollama"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Verify connection
	if err := dbClient.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	log.Printf("Connected to MongoDB at %s", cfg.MongoURI)

	// Initialize dependencies
	sessionStore := database.NewSessionStore(dbClient, "teachme")
	ollamaClient := ollama.NewClient(cfg.OllamaHost)
	chatHandler := handlers.NewChatHandler(sessionStore, ollamaClient)
	sessionHandler := handlers.NewSessionHandler(sessionStore)

	// Setup routes
	http.Handle("/chat", chatHandler)
	http.Handle("/sessions", sessionHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on port %s... (Ollama Host: %s)", cfg.ServerPort, cfg.OllamaHost)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
