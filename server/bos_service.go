package server

import (
	"context"
	"errors"
	"io"
	"net"
	"os"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

// BOSRouter is the interface that defines the entrypoint to the BOS service.
type BOSRouter interface {
	// Route unmarshalls the SNAC frame header from the reader stream to
	// determine which food group to route to. The remainder of the reader
	// stream is passed on to the food group routers for the final SNAC body
	// extraction. Each response sent to the client via the writer stream
	// increments the sequence number.
	Route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32) error
}

// BOSService provides client connection lifecycle management for the BOS
// service.
type BOSService struct {
	AuthHandler
	BOSRouter
	config.Config
	OServiceBOSRouter
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to BOSRouter.
func (rt BOSService) Start() {
	addr := config.Address("", rt.Config.BOSPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		rt.Logger.Error("unable to bind BOS server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting BOS service", "addr", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			rt.Logger.Error(err.Error())
			continue
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "ip", conn.RemoteAddr().String())
		rt.Logger.DebugContext(ctx, "accepted connection")
		go func() {
			if err := rt.handleNewConnection(ctx, conn); err != nil {
				rt.Logger.Info("user session failed", "err", err.Error())
			}
		}()
	}
}

func (rt BOSService) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
	seq := uint32(100)

	flap, err := flapSignonHandshake(rwc, &seq)
	if err != nil {
		return err
	}

	var ok bool
	sessionID, ok := flap.Slice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get session id from payload")
	}

	sess, err := rt.RetrieveBOSSession(string(sessionID))
	if err != nil {
		return err
	}
	if sess == nil {
		return errors.New("session not found")
	}

	defer func() {
		sess.Close()
		rwc.Close()
		if err := rt.Signout(ctx, sess); err != nil {
			rt.Logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
		}
	}()

	ctx = context.WithValue(ctx, "screenName", sess.ScreenName())

	msg := rt.WriteOServiceHostOnline()
	if err := sendSNAC(msg.Frame, msg.Body, &seq, rwc); err != nil {
		return err
	}

	fnClientReqHandler := func(ctx context.Context, r io.Reader, w io.Writer, seq *uint32) error {
		return rt.Route(ctx, sess, r, w, seq)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, seq *uint32) error {
		return sendSNAC(msg.Frame, msg.Body, seq, w)
	}
	dispatchIncomingMessages(ctx, sess, seq, rwc, rt.Logger, fnClientReqHandler, fnAlertHandler)
	return nil
}
