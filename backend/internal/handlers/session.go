package handlers

import (
"encoding/json"
"net/http"

"teachme/internal/types"
)

type SessionHandler struct {
	store types.SessionStore
}

func NewSessionHandler(store types.SessionStore) *SessionHandler {
	return &SessionHandler{store: store}
}

func (h *SessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		return
	}

	sessionID := r.URL.Query().Get("id")

	switch r.Method {
	case "GET":
		h.handleGet(w, r, sessionID)
	case "DELETE":
		h.handleDelete(w, r, sessionID)
	case "PATCH":
		h.handlePatch(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SessionHandler) handleGet(w http.ResponseWriter, r *http.Request, sessionID string) {
	w.Header().Set("Content-Type", "application/json")

	if sessionID != "" {
		session, err := h.store.GetSession(r.Context(), sessionID)
		if err != nil {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(session)
		return
	}

	sessions, err := h.store.GetAllSessions(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch sessions", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		http.Error(w, "Failed to encode sessions", http.StatusInternalServerError)
	}
}

func (h *SessionHandler) handleDelete(w http.ResponseWriter, r *http.Request, sessionID string) {
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteSession(r.Context(), sessionID); err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SessionHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.ID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	session, err := h.store.GetSession(r.Context(), input.ID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	session.Title = input.Title
	if err := h.store.Save(r.Context(), session); err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
