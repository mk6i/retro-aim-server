package oscar

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		logger:      slog.Default(),
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

type fakeConn struct {
	net.Conn // embed the real connection
	local    net.Addr
	remote   net.Addr
}

func (f fakeConn) RemoteAddr() net.Addr { return f.remote }

func TestServer_routeConnection(t *testing.T) {
	sess := state.NewSession()

	clientConn, serverConn := net.Pipe()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	clientFake := fakeConn{
		Conn:   serverConn,
		local:  addr,
		remote: addr,
	}

	go func() {
		// < receive FLAPSignonFrame
		flap := wire.FLAPFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flap, clientConn))
		flapSignonFrame := wire.FLAPSignonFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flapSignonFrame, bytes.NewBuffer(flap.Payload)))

		// > send FLAPSignonFrame
		flapSignonFrame = wire.FLAPSignonFrame{
			FLAPVersion: 1,
		}
		flapSignonFrame.Append(wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, []byte("the-cookie")))
		buf := &bytes.Buffer{}
		assert.NoError(t, wire.MarshalBE(flapSignonFrame, buf))
		flap = wire.FLAPFrame{
			StartMarker: 42,
			FrameType:   wire.FLAPFrameSignon,
			Payload:     buf.Bytes(),
		}
		assert.NoError(t, wire.MarshalBE(flap, clientConn))

		flapc := wire.NewFlapClient(0, clientConn, clientConn)

		// < receive SNAC_0x01_0x03_OServiceHostOnline
		frame := wire.SNACFrame{}
		body := wire.SNAC_0x01_0x03_OServiceHostOnline{}
		assert.NoError(t, flapc.ReceiveSNAC(&frame, &body))

		// send the first request that should get relayed to BOSRouter.Handle
		frame = wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceClientOnline,
		}
		assert.NoError(t, flapc.SendSNAC(frame, struct{}{}))
		assert.NoError(t, clientConn.Close())
	}()

	wg := &sync.WaitGroup{}

	authService := newMockAuthService(t)
	authService.EXPECT().
		RegisterBOSSession(mock.Anything, state.ServerCookie{Service: wire.BOS}).
		Return(sess, nil)
	wg.Add(1)
	authService.EXPECT().
		Signout(mock.Anything, sess).
		Run(func(ctx context.Context, s *state.Session) {
			defer wg.Done()
		})

	authService.EXPECT().
		CrackCookie(mock.Anything).
		Return(state.ServerCookie{Service: wire.BOS}, nil)

	onlineNotifier := newMockOnlineNotifier(t)
	onlineNotifier.EXPECT().
		HostOnline(mock.Anything).
		Return(wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceHostOnline,
			},
			Body: wire.SNAC_0x01_0x03_OServiceHostOnline{},
		})

	buddyListRegistry := newMockBuddyListRegistry(t)
	buddyListRegistry.EXPECT().
		RegisterBuddyList(mock.Anything, mock.Anything).
		Return(nil)
	buddyListRegistry.EXPECT().
		UnregisterBuddyList(mock.Anything, mock.Anything).
		Return(nil)

	departureNotifier := newMockDepartureNotifier(t)
	departureNotifier.EXPECT().
		BroadcastBuddyDeparted(mock.Anything, mock.Anything).
		Return(nil)

	chatSessionManager := newMockChatSessionManager(t)
	chatSessionManager.EXPECT().
		RemoveUserFromAllChats(mock.Anything)

	wg.Add(1)
	handler := func(ctx context.Context, serverType uint16, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter, connectHere string) error {
		defer wg.Done()
		return nil
	}

	rt := oscarServer{
		AuthService:        authService,
		Handler:            handler,
		Logger:             slog.Default(),
		OnlineNotifier:     onlineNotifier,
		BuddyListRegistry:  buddyListRegistry,
		ChatSessionManager: chatSessionManager,
		DepartureNotifier:  departureNotifier,
	}
	assert.NoError(t, rt.routeConnection(context.Background(), clientFake, ""))

	wg.Wait()
}
