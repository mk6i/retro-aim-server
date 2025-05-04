package oscar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// OnlineNotifier returns a OServiceHostOnline SNAC that is sent to the client
// at the beginning of the protocol sequence which lists all food groups
// managed by the server.
type OnlineNotifier interface {
	HostOnline() wire.SNACMessage
}

// BuddyListRegistry is the interface for keeping track of users with active
// buddy lists. Once registered, a user becomes visible to other users' buddy
// lists and vice versa.
type BuddyListRegistry interface {
	ClearBuddyListRegistry(ctx context.Context) error
	RegisterBuddyList(ctx context.Context, user state.IdentScreenName) error
	UnregisterBuddyList(ctx context.Context, user state.IdentScreenName) error
}

// DepartureNotifier is the interface for sending buddy departure notifications
// when a client disconnects.
type DepartureNotifier interface {
	BroadcastBuddyDeparted(ctx context.Context, sess *state.Session) error
}

// ChatSessionManager is the interface for closing chat sessions
// when a client disconnects.
type ChatSessionManager interface {
	RemoveUserFromAllChats(user state.IdentScreenName)
}

// RateLimitUpdater provides rate limit updates for subscribed rate limit classes.
type RateLimitUpdater interface {
	RateLimitUpdates(ctx context.Context, sess *state.Session, now time.Time) []wire.SNACMessage
}

// BOSServer provides client connection lifecycle management for the BOS
// service.
type BOSServer struct {
	AuthService
	BuddyListRegistry
	DepartureNotifier
	Handler
	ListenAddr string
	Logger     *slog.Logger
	OnlineNotifier
	config.Config
	ChatSessionManager *state.InMemoryChatSessionManager
	RateLimitUpdater
	wire.SNACRateLimits
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to BOSRouter.
func (rt BOSServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", rt.ListenAddr)
	if err != nil {
		return fmt.Errorf("unable to start BOS server: %w", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	rt.Logger.Info("starting server", "listen_host", rt.ListenAddr, "oscar_host", rt.Config.OSCARHost)

	if rt.BuddyListRegistry != nil { // nil check is a hack until server refactor
		if err = rt.BuddyListRegistry.ClearBuddyListRegistry(ctx); err != nil {
			return fmt.Errorf("unable to clear client-side buddy list: %s", err.Error())
		}
	}

	wg := sync.WaitGroup{}
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			rt.Logger.Error("accept failed", "err", err.Error())
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			connCtx := context.WithValue(ctx, "ip", conn.RemoteAddr().String())
			rt.Logger.DebugContext(connCtx, "accepted connection")
			if err := rt.handleNewConnection(connCtx, conn); err != nil {
				rt.Logger.Info("user session failed", "err", err.Error())
			}
		}()
	}

	if !waitForShutdown(&wg) {
		rt.Logger.Error("shutdown complete, but connections didn't close cleanly")
	} else {
		rt.Logger.Info("shutdown complete")
	}

	return nil
}

// waitForShutdown returns when either the wg completes or 5 seconds has
// passed. This is a temporary hack to ensure that the server shuts down even
// if all the TCP connections do not drain. Return true if the shutdown is
// clean.
func waitForShutdown(wg *sync.WaitGroup) bool {
	ch := make(chan struct{})

	go func() {
		wg.Wait() // goroutine leak if wg never completes
		close(ch)
	}()

	select {
	case <-ch:
		return true
	case <-time.After(time.Second * 5):
		return false
	}
}

func (rt BOSServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
	defer func() {
		rwc.Close()
	}()

	flapc := wire.NewFlapClient(100, rwc, rwc)

	if err := flapc.SendSignonFrame(nil); err != nil {
		return err
	}
	flap, err := flapc.ReceiveSignonFrame()
	if err != nil {
		return err
	}

	authCookie, ok := flap.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get session id from payload")
	}

	sess, err := rt.RegisterBOSSession(ctx, authCookie)
	if err != nil {
		return err
	}
	if sess == nil {
		return errors.New("session not found")
	}

	if rt.BuddyListRegistry != nil { // nil check is a hack until server refactor
		// todo should this check be below defer()?
		if err := rt.BuddyListRegistry.RegisterBuddyList(ctx, sess.IdentScreenName()); err != nil {
			return fmt.Errorf("unable to init buddy list: %w", err)
		}
	}

	defer func() {
		sess.Close()
		if rt.DepartureNotifier != nil {
			if err := rt.DepartureNotifier.BroadcastBuddyDeparted(ctx, sess); err != nil {
				rt.Logger.ErrorContext(ctx, "error sending buddy departure notifications", "err", err.Error())
			}
		}
		if rt.BuddyListRegistry != nil { // nil check is a hack until server refactor
			// buddy list must be cleared before session is closed, otherwise
			// there will be a race condition that could cause the buddy list
			// be prematurely deleted.
			if err := rt.BuddyListRegistry.UnregisterBuddyList(ctx, sess.IdentScreenName()); err != nil {
				rt.Logger.ErrorContext(ctx, "error removing buddy list entry", "err", err.Error())
			}
		}
		if rt.ChatSessionManager != nil {
			rt.ChatSessionManager.RemoveUserFromAllChats(sess.IdentScreenName())
		}
		rt.Signout(ctx, sess)
	}()

	ctx = context.WithValue(ctx, "screenName", sess.IdentScreenName())

	msg := rt.OnlineNotifier.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	remoteAddr, ok := ctx.Value("ip").(string)
	if ok {
		ip, err := netip.ParseAddrPort(remoteAddr)
		if err != nil {
			return errors.New("unable to parse ip addr")
		}
		sess.SetRemoteAddr(&ip)
	}

	return dispatchIncomingMessages(ctx, sess, flapc, rwc, rt.Logger, rt.Handler, rt.RateLimitUpdater, rt.SNACRateLimits)
}
