package oscar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/netip"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type OSCARServer struct {
	AuthService
	BuddyListRegistry
	ChatSessionManager *state.InMemoryChatSessionManager
	DepartureNotifier
	Logger *slog.Logger
	OnlineNotifier
	Handler
	RateLimitUpdater
	wire.SNACRateLimits
}

func (s OSCARServer) Start(ctx context.Context) error {

	if err := s.BuddyListRegistry.ClearBuddyListRegistry(ctx); err != nil {
		return fmt.Errorf("unable to clear client-side buddy list: %s", err.Error())
	}

	return nil
}

func (s OSCARServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
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

	fg, sess, err := s.GetSession(ctx, authCookie)
	if sess == nil {
		return errors.New("session not found")
	}

	ctx = context.WithValue(ctx, "screenName", sess.IdentScreenName())

	switch fg {
	case 1: // BOS
		if err := s.BuddyListRegistry.RegisterBuddyList(ctx, sess.IdentScreenName()); err != nil {
			return fmt.Errorf("unable to init buddy list: %w", err)
		}

		defer func() {
			sess.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := s.DepartureNotifier.BroadcastBuddyDeparted(ctx, sess); err != nil {
				s.Logger.ErrorContext(ctx, "error sending buddy departure notifications", "err", err.Error())
			}
			// buddy list must be cleared before session is closed, otherwise
			// there will be a race condition that could cause the buddy list
			// be prematurely deleted.
			if err := s.BuddyListRegistry.UnregisterBuddyList(ctx, sess.IdentScreenName()); err != nil {
				s.Logger.ErrorContext(ctx, "error removing buddy list entry", "err", err.Error())
			}
			s.ChatSessionManager.RemoveUserFromAllChats(sess.IdentScreenName())
			s.Signout(ctx, sess)
		}()
		remoteAddr, ok := ctx.Value("ip").(string)
		if ok {
			ip, err := netip.ParseAddrPort(remoteAddr)
			if err != nil {
				return errors.New("unable to parse ip addr")
			}
			sess.SetRemoteAddr(&ip)
		}
	case 2: // Chat
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			s.SignoutChat(ctx, sess)
		}()

	}

	msg := s.OnlineNotifier.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	return dispatchIncomingMessages(ctx, sess, flapc, rwc, s.Logger, s.Handler, s.RateLimitUpdater, s.SNACRateLimits)
}
