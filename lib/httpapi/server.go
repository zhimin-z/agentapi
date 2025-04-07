package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/openagent/lib/logctx"
	st "github.com/coder/openagent/lib/screentracker"
	"github.com/coder/openagent/lib/termexec"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/xerrors"
)

// Server represents the HTTP server
type Server struct {
	router       chi.Router
	api          huma.API
	port         int
	srv          *http.Server
	mu           sync.RWMutex
	logger       *slog.Logger
	conversation *st.Conversation
}

// NewServer creates a new server instance
func NewServer(ctx context.Context, process *termexec.Process, port int) *Server {
	router := chi.NewMux()
	
	// Setup CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	router.Use(corsMiddleware.Handler)
	
	api := humachi.New(router, huma.DefaultConfig("OpenAgent API", "0.1.0"))
	conversation := st.NewConversation(ctx, st.ConversationConfig{
		AgentIO: process,
		GetTime: func() time.Time {
			return time.Now()
		},
		SnapshotInterval:      1 * time.Second,
		ScreenStabilityLength: 2 * time.Second,
	})
	conversation.StartSnapshotLoop(ctx)

	s := &Server{
		router:       router,
		api:          api,
		port:         port,
		conversation: conversation,
		logger:       logctx.From(ctx),
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := s.conversation.Status()
	resp := &StatusResponse{}

	switch status {
	case st.ConversationStatusInitializing:
		resp.Body.Status = "running"
	case st.ConversationStatusStable:
		resp.Body.Status = "waiting_for_input"
	case st.ConversationStatusChanging:
		resp.Body.Status = "running"
	default:
		return nil, xerrors.Errorf("unknown conversation status: %s", status)
	}

	return resp, nil
}

// getMessages handles GET /messages
func (s *Server) getMessages(ctx context.Context, input *struct{}) (*MessagesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := &MessagesResponse{}
	resp.Body.Messages = make([]Message, len(s.conversation.Messages()))
	for i, msg := range s.conversation.Messages() {
		resp.Body.Messages[i] = Message{
			Role:    string(msg.Role),
			Content: msg.Message,
		}
	}

	return resp, nil
}

// createMessage handles POST /message
func (s *Server) createMessage(ctx context.Context, input *MessageRequest) (*MessageResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.conversation.SendMessage(FormatClaudeCodeMessage(input.Body.Content)...); err != nil {
		return nil, xerrors.Errorf("failed to send message: %w", err)
	}

	resp := &MessageResponse{}
	resp.Body.Ok = true

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
