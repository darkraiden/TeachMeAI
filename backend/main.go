package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"teachme/internal/config"
	"teachme/internal/database"
	"teachme/internal/handlers"
	"teachme/internal/middleware"
	"teachme/internal/ollama"
	"teachme/internal/router"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Setup structured logging (JSON) with application name
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("app", "teachme-backend")
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	// Load banned words
	if err := middleware.LoadBannedWords("banned_words.txt"); err != nil {
		slog.Warn("Failed to load banned words", "error", err, "component", "safety")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}

	// Verify connection
	if err := dbClient.Ping(ctx, nil); err != nil {
		slog.Error("Failed to ping MongoDB", "error", err)
		os.Exit(1)
	}

	slog.Info("Connected to MongoDB", "uri", cfg.MongoURI)

	// Initialize dependencies
	sessionStore := database.NewSessionStore(dbClient, "teachme")
	ollamaClient := ollama.NewClient(cfg.OllamaHost)
	chatHandler := handlers.NewChatHandler(sessionStore, ollamaClient)
	sessionHandler := handlers.NewSessionHandler(sessionStore)

	// Setup routes
	r := router.NewRouter(chatHandler, sessionHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	slog.Info("Server starting", "port", cfg.ServerPort, "ollama_host", cfg.OllamaHost)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
