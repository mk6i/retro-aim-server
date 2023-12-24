package server

import (
	"context"
	"io"
	"net"
	"os"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
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
	Config
	OServiceBOSRouter
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to BOSRouter.
func (rt BOSService) Start() {
	addr := Address("", rt.Config.BOSPort)
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
		go rt.handleNewConnection(ctx, conn)
	}
}

func (rt BOSService) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) {
	seq := uint32(100)

	flap, err := SendAndReceiveSignonFrame(rwc, &seq)
	if err != nil {
		rt.Logger.ErrorContext(ctx, "some error", "err", err.Error())
		return
	}

	var ok bool
	sessionID, ok := flap.Slice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		rt.Logger.ErrorContext(ctx, "unable to get session id from payload")
		return
	}

	sess, err := rt.RetrieveBOSSession(string(sessionID))
	if err != nil {
		rt.Logger.ErrorContext(ctx, "unable retrieve session", "err", err.Error())
		return
	}
	if sess == nil {
		rt.Logger.InfoContext(ctx, "session not found", "err", err.Error())
		return
	}

	defer sess.Close()
	defer rwc.Close()

	go func() {
		<-sess.Closed()
		if err := rt.Signout(ctx, sess); err != nil {
			rt.Logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
		}
	}()

	ctx = context.WithValue(ctx, "screenName", sess.ScreenName())

	msg := rt.WriteOServiceHostOnline()
	if err := sendSNAC(msg.Frame, msg.Body, &seq, rwc); err != nil {
		rt.Logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
		return
	}

	fnClientReqHandler := func(ctx context.Context, r io.Reader, w io.Writer, seq *uint32) error {
		return rt.Route(ctx, sess, r, w, seq)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, seq *uint32) error {
		return sendSNAC(msg.Frame, msg.Body, seq, w)
	}
	dispatchIncomingMessages(ctx, sess, seq, rwc, rt.Logger, fnClientReqHandler, fnAlertHandler)
}
