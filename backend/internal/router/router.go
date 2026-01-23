package router

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"teachme/internal/handlers"
	"teachme/internal/middleware"
)

// NewRouter initializes the router and setup routes
func NewRouter(chatHandler *handlers.ChatHandler, sessionHandler *handlers.SessionHandler) *mux.Router {
	r := mux.NewRouter()

	// Setup routes
	// Handlers handle method validation and CORS internally, so we don't restrict methods here yet.
	r.Handle("/chat", middleware.SafetyMiddleware(chatHandler))
	r.Handle("/sessions", sessionHandler)
	r.Handle("/metrics", promhttp.Handler())

	return r
}
