package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"teachme/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockSessionStore
type MockSessionStore struct {
	mock.Mock
}

func (m *MockSessionStore) GetOrCreate(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionStore) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionStore) Save(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionStore) GetAllSessions(ctx context.Context) ([]*models.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

// MockAIClient
type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) Chat(messages []models.Message, systemPrompt string) (string, error) {
	args := m.Called(messages, systemPrompt)
	return args.String(0), args.Error(1)
}

func TestChatHandler_ServeHTTP(t *testing.T) {
	// Setup
	mockStore := new(MockSessionStore)
	mockClient := new(MockAIClient)
	handler := NewChatHandler(mockStore, mockClient)

	// Test Case 1: Successful Chat
	t.Run("Successful Chat", func(t *testing.T) {
sessionID := primitive.NewObjectID().Hex()
		reqBody := models.ChatRequest{
			Message:   "Hello AI",
			SessionID: sessionID,
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/chat", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		// Mock Expectations
		expectedSession := &models.Session{
			ID:       primitive.NewObjectID(),
			Messages: []models.Message{},
		}
		
		// Note: assert.Anything for context because context changes per request
		mockStore.On("GetOrCreate", mock.Anything, sessionID).Return(expectedSession, nil)
		mockClient.On("Chat", mock.AnythingOfType("[]models.Message"), mock.AnythingOfType("string")).Return("Hello Human", nil)
		mockStore.On("Save", mock.Anything, expectedSession).Return(nil)

		// Execute
		handler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp models.ChatResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Hello Human", resp.Reply)
		
		mockStore.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	// Test Case 2: AI Error
	t.Run("AI Error", func(t *testing.T) {
// Reset mocks for clean state
mockStore = new(MockSessionStore)
mockClient = new(MockAIClient)
handler = NewChatHandler(mockStore, mockClient)

reqBody := models.ChatRequest{Message: "Fail me"}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest("POST", "/chat", bytes.NewBuffer(jsonBody))
w := httptest.NewRecorder()

		expectedSession := &models.Session{ID: primitive.NewObjectID()}

		mockStore.On("GetOrCreate", mock.Anything, "").Return(expectedSession, nil)
		mockClient.On("Chat", mock.Anything, mock.Anything).Return("", errors.New("AI offline"))

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)
	})
}

func (m *MockSessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}
