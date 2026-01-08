package database

import (
	"context"
	"time"

	"teachme/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SessionStore struct {
	collection *mongo.Collection
}

func NewSessionStore(client *mongo.Client, dbName string) *SessionStore {
	return &SessionStore{
		collection: client.Database(dbName).Collection("sessions"),
	}
}

// GetOrCreate retrieves an existing session or creates a new one
func (s *SessionStore) GetOrCreate(ctx context.Context, sessionID string) (*models.Session, error) {
	if sessionID != "" {
		// Try to fetch existing session
		objID, err := primitive.ObjectIDFromHex(sessionID)
		if err == nil {
			var session models.Session
			err = s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&session)
			if err == nil {
				return &session, nil
			}
		}
	}

	// Create new session
	session := &models.Session{
		Messages:  []models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}

	session.ID = result.InsertedID.(primitive.ObjectID)
	return session, nil
}

// GetSession retrieves an existing session by ID or returns error
func (s *SessionStore) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	objID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, err
	}

	var session models.Session
	err = s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Save persists a session to the database
func (s *SessionStore) Save(ctx context.Context, session *models.Session) error {
	session.UpdatedAt = time.Now()

	_, err := s.collection.ReplaceOne(
		ctx,
		bson.M{"_id": session.ID},
		session,
	)

	return err
}

// GetAllSessions retrieves all sessions sorted by update time descending
func (s *SessionStore) GetAllSessions(ctx context.Context) ([]*models.Session, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	
	cursor, err := s.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*models.Session
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// DeleteSession removes a session from the database
func (s *SessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	objID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return err
	}

	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
