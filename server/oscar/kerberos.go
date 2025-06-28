package oscar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func NewKerberosServer(
	cfg config.Config,
	logger *slog.Logger,
	authService AuthService,
) *KerberosServer {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /", func(writer http.ResponseWriter, request *http.Request) {
		postKerberosHandler(writer, request, authService, logger)
	})

	return &KerberosServer{
		Server: http.Server{
			Addr:    net.JoinHostPort("", cfg.KerberosPort),
			Handler: mux,
		},
		Logger: logger,
	}
}

// KerberosServer hosts an HTTP endpoint capable of handling AIM-style Kerberos
// authentication. The messages are structured as SNACs transmitted over HTTP.
type KerberosServer struct {
	http.Server
	Logger *slog.Logger
}

func (s *KerberosServer) Start(ctx context.Context) error {
	ch := make(chan error)

	go func() {
		s.Logger.Info("starting kerberos server", "addr", s.Addr)
		if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			ch <- fmt.Errorf("unable to start kerberos server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-ch:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		s.Logger.Error("unable to shutdown kerberos server", "err", err.Error())
	}
	return nil
}

// postKerberosHandler handles AIM-style Kerberos authentication for AIM 6.0+.
func postKerberosHandler(w http.ResponseWriter, r *http.Request, authService AuthService, logger *slog.Logger) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read HTTP body", http.StatusBadRequest)
		return
	}
	reader := bytes.NewReader(b)

	var header wire.SNACFrame
	if err := wire.UnmarshalBE(&header, reader); err != nil {
		http.Error(w, "unable to read kerberos login SNAC header", http.StatusBadRequest)
		return
	}
	if header.FoodGroup != wire.Kerberos && header.FoodGroup != wire.KerberosLoginRequest {
		http.Error(w, "unexpected SNAC type", http.StatusBadRequest)
		return
	}

	var body wire.SNAC_0x050C_0x0002_KerberosLoginRequest
	if err := wire.UnmarshalBE(&body, reader); err != nil {
		http.Error(w, "unable to read kerberos login SNAC body", http.StatusBadRequest)
		return
	}

	response, err := authService.KerberosLogin(r.Context(), body, state.NewStubUser)
	if err != nil {
		logger.Error("authService.KerberosLogin", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-snac")

	if err := wire.MarshalBE(response, w); err != nil {
		logger.Error("unable to marshal SNAC response", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
