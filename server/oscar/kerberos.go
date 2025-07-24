package oscar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"golang.org/x/sync/errgroup"
)

func NewKerberosServer(listeners []config.Listener, logger *slog.Logger, authService AuthService) *KerberosServer {
	servers := make([]*http.Server, 0, len(listeners))
	for _, l := range listeners {

		mux := http.NewServeMux()

		mux.HandleFunc("POST /", func(writer http.ResponseWriter, request *http.Request) {
			postKerberosHandler(writer, request, authService, logger)
		})

		servers = append(servers, &http.Server{
			Addr:    l.KerberosListenAddress,
			Handler: mux,
		})
	}

	return &KerberosServer{
		Servers: servers,
		Logger:  logger,
	}
}

// KerberosServer hosts an HTTP endpoint capable of handling AIM-style Kerberos
// authentication. The messages are structured as SNACs transmitted over HTTP.
type KerberosServer struct {
	Servers []*http.Server
	Logger  *slog.Logger
}

func (s *KerberosServer) Start(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)

	for _, server := range s.Servers {
		errGroup.Go(func() error {
			s.Logger.Info("starting kerberos server", "addr", server.Addr)
			if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("unable to start kerberos server: %w", err)
			}
			return nil
		})

		errGroup.Go(func() error {
			<-ctx.Done()

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				s.Logger.Error("unable to shutdown kerberos server", "err", err.Error())
			}

			return nil
		})
	}

	return errGroup.Wait()
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
