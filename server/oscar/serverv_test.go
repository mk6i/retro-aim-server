package oscar

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/stretchr/testify/assert"
)

func TestServer_ListenAndServeAndShutdown(t *testing.T) {
	var mu sync.Mutex
	var received []string

	var wg sync.WaitGroup

	cfg := []config.Listener{
		{
			BOSListenAddress:  ":1000",
			BOSAdvertisedHost: "localhost",
		},
		{
			BOSListenAddress:  ":2000",
			BOSAdvertisedHost: "localhost",
		},
		{
			BOSListenAddress:  ":3000",
			BOSAdvertisedHost: "localhost",
		},
	}
	responses := []string{"hello1", "hello2", "hello2"}

	server := &Server{
		listenerCfg: cfg,
		conns:       make(map[net.Conn]struct{}),
		closed:      make(chan struct{}),
		handler: func(ctx context.Context, conn net.Conn, advertisedHost string) error {
			for {
				r := bufio.NewReader(conn)
				line, err := r.ReadString('\n')
				if err != nil {
					break
				}
				mu.Lock()
				received = append(received, strings.TrimSpace(line))
				mu.Unlock()
				wg.Done()
			}
			return nil
		},
	}
	server.shutdownCtx, server.shutdownCancel = context.WithCancel(context.Background())

	shutdownCh := make(chan struct{})
	go func() {
		defer close(shutdownCh)
		assert.NoError(t, server.ListenAndServe())
	}()

	for i := 0; i < len(cfg); i++ {
		wg.Add(1)
		// Connect and send message
		conn, err := net.Dial("tcp", "localhost"+cfg[i].BOSListenAddress)
		assert.NoError(t, err)

		_, err = conn.Write([]byte(responses[i] + "\n"))
		assert.NoError(t, err)
	}

	wg.Wait()

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := server.Shutdown(ctx)
	assert.NoError(t, err)

	<-shutdownCh

	// Check what was received
	mu.Lock()
	defer mu.Unlock()
	assert.ElementsMatch(t, received, responses)
}
