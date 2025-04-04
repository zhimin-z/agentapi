package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Server represents the HTTP server
type Server struct {
	router   chi.Router
	api      huma.API
	port     int
	srv      *http.Server
	messages []Message
	mu       sync.RWMutex
}

// NewServer creates a new server instance
func NewServer(port int) *Server {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("OpenAgent API", "1.0.0"))

	s := &Server{
		router:   router,
		api:      api,
		port:     port,
		messages: []Message{},
	}

	// Register API routes
	s.registerRoutes()

	return s
}

// registerRoutes sets up all API endpoints
func (s *Server) registerRoutes() {
	// GET /status endpoint
	huma.Get(s.api, "/status", s.getStatus)

	// GET /messages endpoint
	huma.Get(s.api, "/messages", s.getMessages)

	// POST /message endpoint
	huma.Post(s.api, "/message", s.createMessage)
}

// getStatus handles GET /status
func (s *Server) getStatus(ctx context.Context, input *struct{}) (*StatusResponse, error) {
	resp := &StatusResponse{}
	resp.Body.Status = "running"
	return resp, nil
}

// getMessages handles GET /messages
func (s *Server) getMessages(ctx context.Context, input *struct{}) (*MessagesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := &MessagesResponse{}
	resp.Body.Messages = make([]Message, len(s.messages))
	copy(resp.Body.Messages, s.messages)

	return resp, nil
}

// createMessage handles POST /message
func (s *Server) createMessage(ctx context.Context, input *MessageRequest) (*MessageResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		ID:      uuid.New().String(),
		Content: input.Body.Content,
	}

	s.messages = append(s.messages, msg)

	resp := &MessageResponse{}
	resp.Body.Message = msg

	return resp, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.srv = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return s.srv.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}
