package toc

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"
	"github.com/stretchr/testify/assert"
)

// ensure correct behavior during global context cancellation (server shutdown)
func TestServer_doIt_serverShutdown(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer wg.Done()
		sv := Server{
			BOSProxy: OSCARProxy{},
			Logger:   slog.Default(),
		}

		serverReader, _ := io.Pipe()

		fc := wire.NewFlapClient(0, serverReader, nil)
		closeConn := func() {
			_ = serverReader.Close()
		}
		sess := newTestSession("me")
		err := sv.doIt(ctx, closeConn, sess, nil, fc)
		assert.ErrorIs(t, err, io.ErrClosedPipe)
	}()

	cancel()
	wg.Wait()
}

// ensure correct behavior when client TCP connection disconnects
func TestServer_doIt_clientReadDisconnect(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	serverReader, _ := io.Pipe()

	go func() {
		defer wg.Done()
		closeConn := func() {
			_ = serverReader.Close()
		}
		sess := newTestSession("me")
		fc := wire.NewFlapClient(0, serverReader, nil)

		sv := Server{
			BOSProxy: OSCARProxy{},
			Logger:   slog.Default(),
		}
		err := sv.doIt(context.Background(), closeConn, sess, nil, fc)
		assert.ErrorIs(t, err, io.ErrClosedPipe)
	}()

	_ = serverReader.Close()
	wg.Wait()
}

// ensure correct behavior when session gets closed by another login
func TestServer_doIt_sessClose(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	sess := newTestSession("me")

	go func() {
		defer wg.Done()

		serverReader, _ := io.Pipe()
		fc := wire.NewFlapClient(0, serverReader, nil)

		closeConn := func() {
			_ = serverReader.Close()
		}

		sv := Server{
			BOSProxy: OSCARProxy{},
			Logger:   slog.Default(),
		}
		err := sv.doIt(context.Background(), closeConn, sess, nil, fc)
		assert.NoError(t, err)
	}()

	sess.Close()
	wg.Wait()
}
