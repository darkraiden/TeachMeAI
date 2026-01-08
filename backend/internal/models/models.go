package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message represents a single chat message
type Message struct {
	Role      string    `bson:"role" json:"role"` // "user" or "assistant"
	Content   string    `bson:"content" json:"content"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// Session represents a conversation session
type Session struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title,omitempty" json:"title,omitempty"`
	Messages  []Message          `bson:"messages" json:"messages"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// ChatRequest represents an incoming chat request from the frontend
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// ChatResponse represents the response sent back to the frontend
type ChatResponse struct {
	Reply     string `json:"reply"`
	SessionID string `json:"session_id"`
}
