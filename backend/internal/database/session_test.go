package database

import (
"context"
"testing"
"time"

"teachme/internal/models"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
"go.mongodb.org/mongo-driver/bson/primitive"
)

// SetupTestDB creates a connection to the test database
func SetupTestDB(t *testing.T) *SessionStore {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to localhost:27017 (exposed by Docker)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	// Verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
	}

	// Use a separate test database
	return NewSessionStore(client, "teachme_test")
}

func TestSessionStore_GetOrCreate(t *testing.T) {
	store := SetupTestDB(t)
	ctx := context.Background()

	t.Run("Create New Session", func(t *testing.T) {
session, err := store.GetOrCreate(ctx, "")
require.NoError(t, err)
assert.NotNil(t, session)
assert.False(t, session.ID.IsZero())
		assert.Empty(t, session.Messages)
	})

	t.Run("Get Existing Session", func(t *testing.T) {
// First create one
newSession, err := store.GetOrCreate(ctx, "")
require.NoError(t, err)

// Then fetch it
existingSession, err := store.GetOrCreate(ctx, newSession.ID.Hex())
		require.NoError(t, err)
		assert.Equal(t, newSession.ID, existingSession.ID)
	})
    
    t.Run("Get Non-Existent Session ID Returns New", func(t *testing.T) {
        randomID := primitive.NewObjectID().Hex()
        session, err := store.GetOrCreate(ctx, randomID)
        require.NoError(t, err)
        assert.NotEqual(t, randomID, session.ID.Hex())
        assert.NotNil(t, session)
    })
}

func TestSessionStore_Save(t *testing.T) {
	store := SetupTestDB(t)
	ctx := context.Background()

	session, err := store.GetOrCreate(ctx, "")
	require.NoError(t, err)

	// Modify session
	msg := models.Message{
		Role:      "user",
		Content:   "test message",
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, msg)

	// Save
	err = store.Save(ctx, session)
	assert.NoError(t, err)

	// Verify persistence
	loadedSession, err := store.GetOrCreate(ctx, session.ID.Hex())
	require.NoError(t, err)
	assert.Len(t, loadedSession.Messages, 1)
	assert.Equal(t, "test message", loadedSession.Messages[0].Content)
}
