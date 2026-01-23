package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"teachme/internal/metrics"
	"teachme/internal/models"
	"teachme/internal/ollama"
	"time"
)

// BannedWords is a list of words that should trigger the safety filter.
var BannedWords []string

// LoadBannedWords loads banned words from a file (one word per line).
func LoadBannedWords(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, strings.ToLower(word))
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	BannedWords = words
	slog.Info("Loaded banned words", "count", len(BannedWords), "path", path)
	return nil
}

// SafetyMiddleware checks for banned words in the usage request AND the AI response.
// It inspects the Request Body for JSON content containing a "message" field.
// It also buffers the response to check for unsafe content before sending it to the client.
func SafetyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Only check POST requests (where data is submitted)
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// --- REQUEST CHECKING ---

		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Middleware error reading body", "error", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Restore the body so the next handler can read it
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Try to decode to verify content
		var req models.ChatRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			// If it's not valid JSON, pass it through.
			next.ServeHTTP(w, r)
			return
		}

		// Check message content against banned words
		if containsBannedWords(req.Message) {
			slog.Info("SAFETY: Blocked request containing unsafe content", "status", "blocked", "source", "input", "message_length", len(req.Message), "input_query", req.Message)
			metrics.RequestCounter.WithLabelValues("blocked", "input").Inc()
			http.Error(w, "I'm sorry, but I can't help with that topic. Let's talk about something else! 🦁", http.StatusBadRequest)
			return
		}

		// --- RESPONSE CHECKING ---

		// Wrap the writer to capture the response
		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record duration
		duration := time.Since(start).Seconds()

		// If the handler caused an error (non-200), just pass it through
		// (unless we want to filter error messages too, but usually not needed)
		if rw.statusCode >= 400 {
			w.WriteHeader(rw.statusCode)
			w.Write(rw.body.Bytes())
			metrics.LatencyHistogram.WithLabelValues("error").Observe(duration)
			return
		}

		// Check response body for banned words
		responseBody := rw.body.String()
		if containsBannedWords(responseBody) {
			slog.Info("SAFETY: Blocked RESPONSE containing unsafe content", "status", "blocked", "source", "output")
			metrics.RequestCounter.WithLabelValues("blocked", "output").Inc()
			metrics.LatencyHistogram.WithLabelValues("blocked").Observe(duration)
			http.Error(w, "I'm sorry, but I cannot show you that answer. 🦁", http.StatusBadGateway)
			return
		}

		// Check if LLM self-refused (detected by the refusal prefix in the response)
		if containsLLMRefusal(responseBody) {
			slog.Info("SAFETY: LLM self-refused to answer", "status", "llm-refused", "source", "output", "input_query", req.Message, "duration", duration)
			metrics.RequestCounter.WithLabelValues("llm-refused", "output").Inc()
			metrics.LatencyHistogram.WithLabelValues("llm-refused").Observe(duration)
			// Still return the response - it's the LLM's polite refusal
			w.WriteHeader(rw.statusCode)
			w.Write(rw.body.Bytes())
			return
		}

		// If safe, write the buffered response
		slog.Info("SAFETY: Request processed successfully", "status", "allowed", "source", "output", "duration", duration)
		metrics.RequestCounter.WithLabelValues("allowed", "output").Inc()
		metrics.LatencyHistogram.WithLabelValues("allowed").Observe(duration)
		w.WriteHeader(rw.statusCode)
		w.Write(rw.body.Bytes())
	})
}

// responseWriterWrapper captures the response body
type responseWriterWrapper struct {
	http.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return len(b), nil
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func containsBannedWords(message string) bool {
	lowerMsg := strings.ToLower(message)
	for _, word := range BannedWords {
		// Use word boundary matching to avoid false positives on substrings
		// \b matches word boundaries (start/end of word)
		pattern := `\b` + regexp.QuoteMeta(word) + `\b`
		matched, err := regexp.MatchString(pattern, lowerMsg)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// containsLLMRefusal checks if the LLM response contains refusal patterns,
// indicating the LLM self-moderated its response due to unsafe content.
func containsLLMRefusal(response string) bool {
	lowerResponse := strings.ToLower(strings.TrimSpace(response))

	for _, pattern := range ollama.LLMRefusalPatterns {
		if strings.Contains(lowerResponse, pattern) {
			return true
		}
	}
	return false
}
