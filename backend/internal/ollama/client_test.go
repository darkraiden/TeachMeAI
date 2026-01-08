package ollama

import (
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"testing"
"time"

"teachme/internal/models"

"github.com/stretchr/testify/assert"
)

func TestClient_Chat(t *testing.T) {
	t.Run("Successfully sends request and parses response", func(t *testing.T) {
// Mock server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// Verify URL
assert.Equal(t, "/api/chat", r.URL.Path)

// Verify Request Body
var req ChatRequest
err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)

			// Check defaults
			assert.Equal(t, DefaultModel, req.Model)
			assert.False(t, req.Stream)
			assert.Len(t, req.Messages, 2) // System + User
			assert.Equal(t, "system", req.Messages[0].Role)
			assert.Equal(t, "user", req.Messages[1].Role)
			assert.Equal(t, "Hello", req.Messages[1].Content)

			// Send Response
			resp := ChatResponse{
				Message: ChatMessage{
					Role:    "assistant",
					Content: "Hi there!",
				},
				Done: true,
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		
		messages := []models.Message{
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
		}

		response, err := client.Chat(messages, DefaultSystemPrompt)
		assert.NoError(t, err)
		assert.Equal(t, "Hi there!", response)
	})

	t.Run("Truncates request history to last 20 messages", func(t *testing.T) {
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
var req ChatRequest
json.NewDecoder(r.Body).Decode(&req)
			
			// System prompt + 20 messages = 21 total
			assert.Len(t, req.Messages, 21)
			
			// Verify the first message (after system) is the 6th message from the original list (index 5)
			// Original list: 0..24 (25 msgs)
			// Last 20 are indices 5..24
			
			// req.Messages[0] is system prompt
			// req.Messages[1] should be messages[5] ("msg-5")
			// req.Messages[20] should be messages[24] ("msg-24")
			
			assert.Equal(t, "msg-5", req.Messages[1].Content)
			assert.Equal(t, "msg-24", req.Messages[20].Content)

			json.NewEncoder(w).Encode(ChatResponse{
Message: ChatMessage{Content: "Response"},
})
		}))
		defer server.Close()

		client := NewClient(server.URL)

		// Create 25 messages
		var messages []models.Message
		for i := 0; i < 25; i++ {
			messages = append(messages, models.Message{
Role:    "user",
Content: fmt.Sprintf("msg-%d", i),
})
		}
		
		_, err := client.Chat(messages, DefaultSystemPrompt)
		assert.NoError(t, err)
	})
}
