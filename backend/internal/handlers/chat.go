package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"teachme/internal/models"
	"teachme/internal/ollama"
	"teachme/internal/types"
)

type ChatHandler struct {
	sessionStore types.SessionStore
	ollamaClient types.AIClient
}

func NewChatHandler(sessionStore types.SessionStore, ollamaClient types.AIClient) *ChatHandler {
	return &ChatHandler{
		sessionStore: sessionStore,
		ollamaClient: ollamaClient,
	}
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use a short timeout for initial session retrieval
	getCtx, getCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer getCancel()

	session, err := h.sessionStore.GetOrCreate(getCtx, req.SessionID)
	if err != nil {
		slog.Error("Session error", "error", err)
		http.Error(w, "Failed to manage session", http.StatusInternalServerError)
		return
	}

	// Add user message to history
	userMsg := models.Message{
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, userMsg)

	// Get AI response (this can take a while, no context timeout here - HTTP client has its own)
	aiResponse, err := h.ollamaClient.Chat(session.Messages, ollama.DefaultSystemPrompt)
	if err != nil {
		slog.Error("Ollama error", "error", err)
		http.Error(w, "Failed to get AI response", http.StatusBadGateway)
		return
	}

	// Save assistant message to history
	assistantMsg := models.Message{
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, assistantMsg)

	// Use a fresh context for saving (the previous one may have expired during Ollama call)
	saveCtx, saveCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer saveCancel()

	// Persist to database
	if err := h.sessionStore.Save(saveCtx, session); err != nil {
		slog.Error("Failed to save session", "error", err)
		// Continue anyway - don't fail the request
	}

	// Respond to frontend
	response := models.ChatResponse{
		Reply:     aiResponse,
		SessionID: session.ID.Hex(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, PATCH, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
