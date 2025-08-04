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
	"golang.org/x/time/rate"
)

func TestServer_ListenAndServeAndShutdown(t *testing.T) {
	var mu sync.Mutex
	var received []string

	var msgWg sync.WaitGroup

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

	server := NewServer(
		nil,
		nil,
		nil,
		nil,
		slog.Default(),
		nil,
		nil,
		nil,
		wire.DefaultSNACRateLimits(),
		nil,
		cfg,
	)

	server.handler = func(ctx context.Context, conn net.Conn, advertisedHost string) error {
		go func() {
			<-ctx.Done()
			_ = conn.Close()
		}()
		for {
			r := bufio.NewReader(conn)
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			mu.Lock()
			received = append(received, strings.TrimSpace(line))
			mu.Unlock()
			msgWg.Done()
		}
		return nil
	}
	server.shutdownCtx, server.shutdownCancel = context.WithCancel(context.Background())

	shutdownCh := make(chan struct{})
	go func() {
		defer close(shutdownCh)
		assert.NoError(t, server.ListenAndServe())
	}()

	for i := 0; i < len(cfg); i++ {
		msgWg.Add(1)
		// Connect and send message
		conn, err := net.Dial("tcp", "localhost"+cfg[i].BOSListenAddress)
		assert.NoError(t, err)

		_, err = conn.Write([]byte(responses[i] + "\n"))
		assert.NoError(t, err)
	}

	msgWg.Wait()

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func TestOscarServer_RouteConnection_Auth_BUCP(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	assert.NoError(t, err)

	clientFake := fakeConn{
		Conn:   serverConn,
		local:  addr,
		remote: addr,
	}

	go func() {
		defer func() {
			_ = clientConn.Close()
		}()

		// < receive FLAPSignonFrame
		flap := wire.FLAPFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flap, clientConn))
		flapSignonFrame := wire.FLAPSignonFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flapSignonFrame, bytes.NewBuffer(flap.Payload)))

		// > send FLAPSignonFrame
		flapSignonFrame = wire.FLAPSignonFrame{
			FLAPVersion: 1,
		}
		buf := &bytes.Buffer{}
		assert.NoError(t, wire.MarshalBE(flapSignonFrame, buf))
		flap = wire.FLAPFrame{
			StartMarker: 42,
			FrameType:   wire.FLAPFrameSignon,
			Payload:     buf.Bytes(),
		}
		assert.NoError(t, wire.MarshalBE(flap, clientConn))

		// > send SNAC_0x17_0x06_BUCPChallengeRequest
		flapc := wire.NewFlapClient(0, clientConn, clientConn)
		frame := wire.SNACFrame{
			FoodGroup: wire.BUCP,
			SubGroup:  wire.BUCPChallengeRequest,
		}
		bodyIn := wire.SNAC_0x17_0x06_BUCPChallengeRequest{}
		assert.NoError(t, flapc.SendSNAC(frame, bodyIn))

		// < receive SNAC_0x17_0x07_BUCPChallengeResponse
		frame = wire.SNACFrame{}
		assert.NoError(t, flapc.ReceiveSNAC(&frame, &wire.SNAC_0x17_0x07_BUCPChallengeResponse{}))
		assert.Equal(t, wire.SNACFrame{FoodGroup: wire.BUCP, SubGroup: wire.BUCPChallengeResponse}, frame)

		// > send keep alive frame (like BSFlite does mid-login)
		assert.NoError(t, flapc.SendKeepAliveFrame())

		// > send SNAC_0x17_0x02_BUCPLoginRequest
		frame = wire.SNACFrame{
			FoodGroup: wire.BUCP,
			SubGroup:  wire.BUCPLoginRequest,
		}
		assert.NoError(t, flapc.SendSNAC(frame, wire.SNAC_0x17_0x02_BUCPLoginRequest{}))

		// < receive SNAC_0x17_0x03_BUCPLoginResponse
		frame = wire.SNACFrame{}
		assert.NoError(t, flapc.ReceiveSNAC(&frame, &wire.SNAC_0x17_0x03_BUCPLoginResponse{}))
		assert.Equal(t, wire.SNACFrame{FoodGroup: wire.BUCP, SubGroup: wire.BUCPLoginResponse}, frame)
	}()

	wg := &sync.WaitGroup{}

	authService := newMockAuthService(t)
	authService.EXPECT().
		BUCPChallenge(matchContext(), mock.Anything, mock.Anything).
		Return(wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.BUCP,
				SubGroup:  wire.BUCPChallengeResponse,
			},
			Body: wire.SNAC_0x17_0x07_BUCPChallengeResponse{},
		}, nil)
	authService.EXPECT().
		BUCPLogin(matchContext(), mock.Anything, mock.Anything, "localhost:5190").
		Return(wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.BUCP,
				SubGroup:  wire.BUCPLoginResponse,
			},
			Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{},
		}, nil)

	rt := oscarServer{
		AuthService:   authService,
		Logger:        slog.Default(),
		IPRateLimiter: NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
	}
	assert.NoError(t, rt.routeConnection(context.Background(), clientFake, "localhost:5190"))

	wg.Wait()
}

func TestOscarServer_RouteConnection_Auth_FLAP(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	assert.NoError(t, err)

	clientFake := fakeConn{
		Conn:   serverConn,
		local:  addr,
		remote: addr,
	}

	go func() {
		defer func() {
			_ = clientConn.Close()
		}()

		// < receive FLAPSignonFrame
		flap := wire.FLAPFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flap, clientConn))
		flapSignonFrame := wire.FLAPSignonFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flapSignonFrame, bytes.NewBuffer(flap.Payload)))

		// > send FLAPSignonFrame with screen name TLV (indicates FLAP auth)
		flapSignonFrame = wire.FLAPSignonFrame{
			FLAPVersion: 1,
		}
		// Add screen name TLV to indicate FLAP authentication
		flapSignonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsScreenName, "testuser"))
		// Add password hash TLV for authentication
		flapSignonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, []byte("password_hash")))
		// Add client identity TLV
		flapSignonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsClientIdentity, "ICQ 2000b"))

		buf := &bytes.Buffer{}
		assert.NoError(t, wire.MarshalBE(flapSignonFrame, buf))
		flap = wire.FLAPFrame{
			StartMarker: 42,
			FrameType:   wire.FLAPFrameSignon,
			Payload:     buf.Bytes(),
		}
		assert.NoError(t, wire.MarshalBE(flap, clientConn))

		// < receive FLAPSignoffFrame with authentication result
		flap = wire.FLAPFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flap, clientConn))
		assert.Equal(t, wire.FLAPFrameSignoff, flap.FrameType)

		// Parse the signoff frame payload to verify authentication response
		signoffTLVs := wire.TLVRestBlock{}
		assert.NoError(t, wire.UnmarshalBE(&signoffTLVs, bytes.NewBuffer(flap.Payload)))
	}()

	wg := &sync.WaitGroup{}

	authService := newMockAuthService(t)
	authService.EXPECT().
		FLAPLogin(matchContext(), mock.Anything, mock.Anything, "localhost:5190").
		Return(wire.TLVRestBlock{
			TLVList: []wire.TLV{
				wire.NewTLVBE(wire.LoginTLVTagsScreenName, "testuser"),
				wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "localhost:5190"),
				wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("auth-cookie")),
			},
		}, nil)

	rt := oscarServer{
		AuthService:   authService,
		Logger:        slog.Default(),
		IPRateLimiter: NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
	}
	assert.NoError(t, rt.routeConnection(context.Background(), clientFake, "localhost:5190"))

	wg.Wait()
}

func TestOscarServer_RouteConnection_BOS(t *testing.T) {
	sess := state.NewSession()

	clientConn, serverConn := net.Pipe()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	assert.NoError(t, err)

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
	handler := func(ctx context.Context, serverType uint16, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter, advertisedHost string) error {
		defer wg.Done()
		return nil
	}

	rt := oscarServer{
		AuthService:        authService,
		SNACHandler:        handler,
		Logger:             slog.Default(),
		OnlineNotifier:     onlineNotifier,
		BuddyListRegistry:  buddyListRegistry,
		ChatSessionManager: chatSessionManager,
		DepartureNotifier:  departureNotifier,
	}
	assert.NoError(t, rt.routeConnection(context.Background(), clientFake, ""))

	wg.Wait()
}

func TestOscarServer_RouteConnection_Chat(t *testing.T) {
	sess := state.NewSession()

	clientConn, serverConn := net.Pipe()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	assert.NoError(t, err)

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
		RegisterChatSession(mock.Anything, state.ServerCookie{Service: wire.Chat}).
		Return(sess, nil)
	wg.Add(1)
	authService.EXPECT().
		SignoutChat(mock.Anything, sess).
		Run(func(ctx context.Context, s *state.Session) {
			defer wg.Done()
		})

	authService.EXPECT().
		CrackCookie(mock.Anything).
		Return(state.ServerCookie{Service: wire.Chat}, nil)

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
	departureNotifier := newMockDepartureNotifier(t)
	chatSessionManager := newMockChatSessionManager(t)

	wg.Add(1)
	handler := func(ctx context.Context, serverType uint16, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter, advertisedHost string) error {
		defer wg.Done()
		return nil
	}

	rt := oscarServer{
		AuthService:        authService,
		SNACHandler:        handler,
		Logger:             slog.Default(),
		OnlineNotifier:     onlineNotifier,
		BuddyListRegistry:  buddyListRegistry,
		ChatSessionManager: chatSessionManager,
		DepartureNotifier:  departureNotifier,
	}
	assert.NoError(t, rt.routeConnection(context.Background(), clientFake, ""))

	wg.Wait()
}

func TestOscarServer_RouteConnection_Admin(t *testing.T) {
	sess := state.NewSession()

	clientConn, serverConn := net.Pipe()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	assert.NoError(t, err)

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
		CrackCookie(mock.Anything).
		Return(state.ServerCookie{Service: wire.Admin}, nil)
	authService.EXPECT().
		RetrieveBOSSession(mock.Anything, state.ServerCookie{Service: wire.Admin}).
		Return(sess, nil)

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
	departureNotifier := newMockDepartureNotifier(t)
	chatSessionManager := newMockChatSessionManager(t)

	wg.Add(1)
	handler := func(ctx context.Context, serverType uint16, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter, advertisedHost string) error {
		defer wg.Done()
		return nil
	}

	rt := oscarServer{
		AuthService:        authService,
		SNACHandler:        handler,
		Logger:             slog.Default(),
		OnlineNotifier:     onlineNotifier,
		BuddyListRegistry:  buddyListRegistry,
		ChatSessionManager: chatSessionManager,
		DepartureNotifier:  departureNotifier,
	}
	assert.NoError(t, rt.routeConnection(context.Background(), clientFake, ""))

	wg.Wait()
}
