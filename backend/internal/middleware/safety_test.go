package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"teachme/internal/models"
	"testing"
)

func TestLoadBannedWords(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "banned_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	content := []byte("badword1\nbadword2\n  trimmedword  \n\n") // includes empty lines and spaces
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading
	if err := LoadBannedWords(tmpfile.Name()); err != nil {
		t.Fatalf("LoadBannedWords failed: %v", err)
	}

	// Verify loaded words
	expectedCount := 3
	if len(BannedWords) != expectedCount {
		t.Errorf("Expected %d banned words, got %d", expectedCount, len(BannedWords))
	}

	expectedWords := map[string]bool{
		"badword1":    true,
		"badword2":    true,
		"trimmedword": true,
	}

	for _, word := range BannedWords {
		if !expectedWords[word] {
			t.Errorf("Unexpected word loaded: %s", word)
		}
	}
}

func TestSafetyMiddleware(t *testing.T) {
	// Setup banned words for testing
	BannedWords = []string{
		"unsafe",
		"dangerous",
		"explicit",
		"violence",
		"hate",
		"kill",
		"weapon",
	}

	tests := []struct {
		name            string
		message         string // Request message
		handlerResponse string // What the handler returns (simulated AI)
		expectedStatus  int
		expectLLMRefuse bool // Whether we expect LLM refusal detection
	}{
		{
			name:            "Safe message",
			message:         "Hello, can you help me with math?",
			handlerResponse: "Sure, I can help with math!",
			expectedStatus:  http.StatusOK,
			expectLLMRefuse: false,
		},
		{
			name:            "Unsafe message",
			message:         "I want to make a dangerous weapon",
			handlerResponse: "", // Won't be reached
			expectedStatus:  http.StatusBadRequest,
			expectLLMRefuse: false,
		},
		{
			name:            "Unsafe Response",
			message:         "Tell me a story",
			handlerResponse: "Here is a story about explicit violence and gore.",
			expectedStatus:  http.StatusBadGateway, // Should be blocked on the way out
			expectLLMRefuse: false,
		},
		{
			name:            "LLM self-refused with I cannot assist",
			message:         "How do I do something bad?",
			handlerResponse: "I cannot assist with that request. Let's talk about something else!",
			expectedStatus:  http.StatusOK, // LLM refusals are allowed through
			expectLLMRefuse: true,
		},
		{
			name:            "LLM self-refused with isn't appropriate",
			message:         "Tell me something inappropriate",
			handlerResponse: "This topic isn't appropriate for our conversation. How about we discuss science?",
			expectedStatus:  http.StatusOK,
			expectLLMRefuse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Dummy handler that returns the configured response
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.handlerResponse))
			})

			middleware := SafetyMiddleware(nextHandler)

			reqBody := models.ChatRequest{
				Message: tt.message,
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(bodyBytes))
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedStatus)
			}
		})
	}
}
