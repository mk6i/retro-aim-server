package webapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/mk6i/retro-aim-server/server/webapi/middleware"
)

func NewServer(listeners []string, logger *slog.Logger, handler Handler, apiKeyValidator middleware.APIKeyValidator) *Server {
	servers := make([]*http.Server, 0, len(listeners))

	// Create authentication middleware
	// TODO: Remove underscore when implementing authenticated endpoints
	_ = middleware.NewAuthMiddleware(apiKeyValidator, logger)

	for _, l := range listeners {
		mux := http.NewServeMux()
		
		// Public endpoint (no auth required for hello world)
		mux.HandleFunc("GET /", handler.GetHelloWorldHandler)
		
		// TODO: Add authenticated Web AIM API endpoints here
		// Example:
		// mux.Handle("GET /aim/startSession", authMiddleware.Authenticate(
		//     authMiddleware.CORSMiddleware(
		//         http.HandlerFunc(handler.StartSessionHandler))))

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
