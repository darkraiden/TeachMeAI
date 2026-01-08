package types

import (
	"context"

	"teachme/internal/models"
)

// SessionStore defines the interface for session storage operations
type SessionStore interface {
	GetOrCreate(ctx context.Context, sessionID string) (*models.Session, error)
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	Save(ctx context.Context, session *models.Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	GetAllSessions(ctx context.Context) ([]*models.Session, error)
}

// AIClient defines the interface for AI interaction
type AIClient interface {
	Chat(messages []models.Message, systemPrompt string) (string, error)
}
