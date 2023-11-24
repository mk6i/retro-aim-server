package server

import (
	"context"
	"errors"
	"io"
	"net"
	"os"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type BOSService struct {
	AlertRouter
	AuthHandler
	BuddyRouter
	ChatNavRouter
	Config
	FeedbagRouter
	ICBMRouter
	LocateRouter
	OServiceBOSRouter
	RouteLogger
}

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
	sess, seq, err := rt.VerifyLogin(rwc)
	if err != nil {
		rt.Logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
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
		return rt.route(ctx, sess, r, w, seq)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, seq *uint32) error {
		return sendSNAC(msg.Frame, msg.Body, seq, w)
	}
	dispatchIncomingMessages(ctx, sess, seq, rwc, rt.Logger, fnClientReqHandler, fnAlertHandler)
}

func (rt BOSService) route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32) error {
	inFrame := oscar.SNACFrame{}
	if err := oscar.Unmarshal(&inFrame, r); err != nil {
		return err
	}

	err := func() error {
		switch inFrame.FoodGroup {
		case oscar.OService:
			return rt.RouteOService(ctx, sess, inFrame, r, w, sequence)
		case oscar.Locate:
			return rt.RouteLocate(ctx, sess, inFrame, r, w, sequence)
		case oscar.Buddy:
			return rt.RouteBuddy(ctx, inFrame, r, w, sequence)
		case oscar.ICBM:
			return rt.RouteICBM(ctx, sess, inFrame, r, w, sequence)
		case oscar.ChatNav:
			return rt.RouteChatNav(ctx, sess, inFrame, r, w, sequence)
		case oscar.Feedbag:
			return rt.RouteFeedbag(ctx, sess, inFrame, r, w, sequence)
		case oscar.BUCP:
			return routeBUCP(ctx)
		case oscar.Alert:
			return rt.RouteAlert(ctx, inFrame)
		default:
			return ErrUnsupportedSubGroup
		}
	}()

	if err != nil {
		rt.logRequestError(ctx, inFrame, err)
		if errors.Is(err, ErrUnsupportedSubGroup) {
			if err1 := sendInvalidSNACErr(inFrame, w, sequence); err1 != nil {
				err = errors.Join(err1, err)
			}
			if rt.Config.FailFast {
				panic(err.Error())
			}
			return nil
		}
	}

	return err
}
