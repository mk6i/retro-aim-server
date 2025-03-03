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

func TestServer_doIt_serverShutdown(t *testing.T) {
	sv := Server{
		BOSProxy: OSCARProxy{},
		Logger:   slog.Default(),
	}

	serverReader, _ := io.Pipe()
	_, serverWriter := io.Pipe()

	fc := wire.NewFlapClient(0, serverReader, serverWriter)

	wg := sync.WaitGroup{}
	wg.Add(1)

	sess := newTestSession("me")

	ctx, cancel := context.WithCancel(context.Background())

	closeConn := func() {
		_ = serverReader.Close()
		_ = serverWriter.Close()
	}

	go func() {
		defer wg.Done()
		cr := NewChatRegistry()

		err := sv.doIt(ctx, closeConn, sess, cr, fc)
		assert.NoError(t, err)
	}()

	cancel()
	wg.Wait()
}

func TestServer_doIt_clientReadDisconnect(t *testing.T) {
	sv := Server{
		BOSProxy: OSCARProxy{},
		Logger:   slog.Default(),
	}

	serverReader, _ := io.Pipe()
	_, serverWriter := io.Pipe()

	fc := wire.NewFlapClient(0, serverReader, serverWriter)

	wg := sync.WaitGroup{}
	wg.Add(1)

	sess := newTestSession("me")

	ctx := context.Background()

	closeConn := func() {
		_ = serverReader.Close()
		_ = serverWriter.Close()
	}

	go func() {
		defer wg.Done()
		cr := NewChatRegistry()

		err := sv.doIt(ctx, closeConn, sess, cr, fc)
		assert.NoError(t, err)
	}()

	_ = serverReader.Close()
	wg.Wait()
}

func TestServer_doIt_sessClose(t *testing.T) {
	sv := Server{
		BOSProxy: OSCARProxy{},
		Logger:   slog.Default(),
	}

	serverReader, _ := io.Pipe()
	_, serverWriter := io.Pipe()

	fc := wire.NewFlapClient(0, serverReader, serverWriter)

	wg := sync.WaitGroup{}
	wg.Add(1)

	sess := newTestSession("me")

	ctx := context.Background()

	go func() {
		defer wg.Done()
		closeConn := func() {
			_ = serverReader.Close()
			_ = serverWriter.Close()
		}
		cr := NewChatRegistry()

		err := sv.doIt(ctx, closeConn, sess, cr, fc)
		assert.NoError(t, err)
	}()

	sess.Close()
	wg.Wait()
}
