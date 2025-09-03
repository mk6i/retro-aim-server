package webapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/mk6i/retro-aim-server/server/webapi/handlers"
	"github.com/mk6i/retro-aim-server/server/webapi/middleware"
	"github.com/mk6i/retro-aim-server/state"
)

func NewServer(listeners []string, logger *slog.Logger, handler Handler, apiKeyValidator middleware.APIKeyValidator, sessionManager *state.WebAPISessionManager) *Server {
	servers := make([]*http.Server, 0, len(listeners))

	// Create authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(apiKeyValidator, logger)

	// Create handlers
	authHandler := &handlers.AuthHandler{
		UserManager: handler.UserManager,
		TokenStore:  handler.TokenStore,
		Logger:      logger,
	}

	sessionHandler := &handlers.SessionHandler{
		SessionManager:   sessionManager,
		OSCARAuthService: handler.AuthService,
		TokenStore:       handler.TokenStore,
		Logger:           logger,
	}

	eventsHandler := &handlers.EventsHandler{
		SessionManager: sessionManager,
		Logger:         logger,
	}

	presenceHandler := &handlers.PresenceHandler{
		SessionManager:   sessionManager,
		SessionRetriever: handler.SessionRetriever,
		FeedbagRetriever: handler.FeedbagRetriever,
		BuddyBroadcaster: handler.BuddyBroadcaster,
		ProfileManager:   handler.ProfileManager,
		Logger:           logger,
	}

	buddyListHandler := &handlers.BuddyListHandler{
		SessionManager: sessionManager,
		FeedbagManager: handler.FeedbagManager,
		Logger:         logger,
	}

	// Phase 2: Messaging handler
	messagingHandler := &handlers.MessagingHandler{
		SessionManager:        sessionManager,
		MessageRelayer:        handler.MessageRelayer,
		OfflineMessageManager: handler.OfflineMessageManager,
		SessionRetriever:      handler.SessionRetriever,
		Logger:                logger,
	}

	for _, l := range listeners {
		mux := http.NewServeMux()

		// Public endpoint (no auth required for hello world)
		mux.HandleFunc("GET /", handler.GetHelloWorldHandler)

		// Authentication endpoint (public - no API key required for user login)
		// Using pattern with explicit method for Go 1.22+ routing
		mux.HandleFunc("POST /auth/clientLogin", func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers for public endpoint
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			authHandler.ClientLogin(w, r)
		})

		// Handle OPTIONS for CORS preflight
		mux.HandleFunc("OPTIONS /auth/clientLogin", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
		})

		// Authenticated Web AIM API endpoints
		// Session management
		mux.Handle("GET /aim/startSession", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(sessionHandler.StartSession))))

		mux.Handle("GET /aim/endSession", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(sessionHandler.EndSession))))

		// Event fetching
		mux.Handle("GET /aim/fetchEvents", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(eventsHandler.FetchEvents))))

		// Presence and buddy list
		mux.Handle("GET /presence/get", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(presenceHandler.GetPresence))))

		mux.Handle("GET /buddylist/addBuddy", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(buddyListHandler.AddBuddy))))

		// Phase 2: Messaging endpoints
		mux.Handle("GET /im/sendIM", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(messagingHandler.SendIM))))

		mux.Handle("GET /im/setTyping", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(messagingHandler.SetTyping))))

		// Phase 2: Presence management endpoints
		mux.Handle("GET /presence/setState", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(presenceHandler.SetState))))

		mux.Handle("GET /presence/setStatus", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(presenceHandler.SetStatus))))

		mux.Handle("GET /presence/setProfile", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(presenceHandler.SetProfile))))

		mux.Handle("GET /presence/getProfile", authMiddleware.Authenticate(
			authMiddleware.CORSMiddleware(
				http.HandlerFunc(presenceHandler.GetProfile))))

		// Phase 2: Presence icon endpoint (no auth required)
		mux.HandleFunc("GET /presence/icon", presenceHandler.Icon)

		servers = append(servers, &http.Server{
			Addr:    l,
			Handler: mux,
		})
	}

	return &Server{
		servers: servers,
		logger:  logger,
	}
}

// Server hosts an HTTP endpoint capable of handling AIM-style Kerberos
// authentication. The messages are structured as SNACs transmitted over HTTP.
type Server struct {
	servers []*http.Server
	logger  *slog.Logger
}

func (s *Server) ListenAndServe() error {
	if len(s.servers) == 0 {
		s.logger.Debug("no webapi listeners defined")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	for _, server := range s.servers {
		g.Go(func() error {
			s.logger.Info("starting server", "addr", server.Addr)
			if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				cancel()
				return fmt.Errorf("unable to start webapi server: %w", err)
			}
			return nil
		})
	}

	return g.Wait()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if len(s.servers) > 0 {
		for _, srv := range s.servers {
			_ = srv.Shutdown(ctx)
		}
		s.logger.Info("shutdown complete")
	}
	return nil
}
