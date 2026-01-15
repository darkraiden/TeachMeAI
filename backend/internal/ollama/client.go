package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"teachme/internal/models"
)

const (
	DefaultModel        = "phi3"
	DefaultSystemPrompt = "You are a kind, patient, and safe AI tutor for children. " +
		"You explain things simply and clearly. " +
		"If asked about harmful, violent, or adult topics, " +
		"you must refuse by starting your response with: " +
		"\"I can't help with that.\" followed by a gentle explanation and suggestion for a safer topic. " +
		"Do not use complicated words."
)

// LLMRefusalPatterns are phrases that indicate the LLM is self-moderating its response
var LLMRefusalPatterns = []string{
	// Direct refusals
	"i can't help with that",
	"i cannot help with that",
	"i'm not able to help",
	"i am not able to help",
	"i can't assist with",
	"i cannot assist with",
	"i'm unable to",
	"i am unable to",
	"i can't provide information",
	"i cannot provide information",
	// Topic redirection phrases
	"this might not be the topic best suited",
	"not be the topic best suited",
	"isn't appropriate",
	"is not appropriate",
	"isn't safe to explore",
	"is not safe to explore",
	"not appropriate or safe",
	// Safe environment phrases
	"i want to make sure that our talks are always positive",
	"make sure we always have a safe environment",
	"let's keep things light-hearted",
	"let's talk about something else",
	"it's best not to think about",
	"instead of discussing",
	"how about we learn",
	"how about learning",
	// Apology + redirect patterns
	"i'm sorry, but discussing",
	"i am sorry, but discussing",
	"i'm sorry, but i can't",
	"i am sorry, but i cannot",
}

// ChatMessage represents a message in the Ollama chat format
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a request to the Ollama chat endpoint
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// ChatResponse represents a response from the Ollama chat endpoint
type ChatResponse struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second, // LLM responses can be slow
		},
	}
}

// Chat sends a chat request to Ollama and returns the response
func (c *Client) Chat(messages []models.Message, systemPrompt string) (string, error) {
	// Build Ollama messages with system prompt first
	ollamaMessages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
	}

	// Add conversation history (limit to last 20 messages to avoid token overflow)
	startIdx := 0
	if len(messages) > 20 {
		startIdx = len(messages) - 20
	}

	for _, msg := range messages[startIdx:] {
		ollamaMessages = append(ollamaMessages, ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Prepare request
	req := ChatRequest{
		Model:    DefaultModel,
		Messages: ollamaMessages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	resp, err := c.client.Post(
		fmt.Sprintf("%s/api/chat", c.baseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w (body: %s)", err, string(body))
	}

	return chatResp.Message.Content, nil
}
