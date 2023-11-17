package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/user"
	"io"
	"log/slog"
	"net"

	"github.com/mkaminski/goaim/oscar"
)

func NewBOSServiceRouter(logger *slog.Logger, cfg Config, fm FeedbagManager, sm SessionManager, cr *ChatRegistry, pm ProfileManager) BOSServiceRouter {
	return BOSServiceRouter{
		AlertRouter:       NewAlertRouter(logger),
		BuddyRouter:       NewBuddyRouter(logger),
		ChatNavRouter:     NewChatNavRouter(logger, cr),
		FeedbagRouter:     NewFeedbagRouter(logger, sm, fm),
		ICBMRouter:        NewICBMRouter(logger, sm, fm),
		LocateRouter:      NewLocateRouter(logger, sm, fm, pm),
		OServiceBOSRouter: NewOServiceRouterForBOS(logger, cfg, fm, sm, cr),
		sm:                sm,
		fm:                fm,
		cfg:               cfg,
		RouteLogger: RouteLogger{
			Logger: logger,
		},
		NewChatSessMgr: func() ChatSessionManager { return user.NewSessionManager(logger) },
	}
}

func NewChatServiceRouter(logger *slog.Logger, cfg Config, fm FeedbagManager, sm SessionManager) ChatServiceRouter {
	return ChatServiceRouter{
		OServiceChatRouter: NewOServiceRouterForChat(logger, cfg, fm, sm),
		ChatRouter:         NewChatRouter(logger),
		cfg:                cfg,
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type BOSServiceRouter struct {
	AlertRouter
	BuddyRouter
	ChatNavRouter
	FeedbagRouter
	ICBMRouter
	LocateRouter
	OServiceBOSRouter
	sm  SessionManager
	fm  FeedbagManager
	cfg Config
	RouteLogger
	NewChatSessMgr func() ChatSessionManager
}

func (rt *BOSServiceRouter) Route(ctx context.Context, sess *user.Session, r io.Reader, w io.Writer, sequence *uint32) error {
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, r); err != nil {
		return err
	}

	err := func() error {
		switch snac.FoodGroup {
		case oscar.OSERVICE:
			return rt.RouteOService(ctx, sess, snac, r, w, sequence)
		case oscar.LOCATE:
			return rt.RouteLocate(ctx, sess, snac, r, w, sequence)
		case oscar.BUDDY:
			return rt.RouteBuddy(ctx, snac, r, w, sequence)
		case oscar.ICBM:
			return rt.RouteICBM(ctx, sess, snac, r, w, sequence)
		case oscar.CHAT_NAV:
			return rt.RouteChatNav(ctx, sess, rt.NewChatSessMgr, snac, r, w, sequence)
		case oscar.FEEDBAG:
			return rt.RouteFeedbag(ctx, sess, snac, r, w, sequence)
		case oscar.BUCP:
			return routeBUCP(ctx)
		case oscar.ALERT:
			return rt.RouteAlert(ctx, snac)
		default:
			return ErrUnsupportedSubGroup
		}
	}()

	if err != nil {
		rt.logRequestError(ctx, snac, err)
		if errors.Is(err, ErrUnsupportedSubGroup) {
			if err1 := sendInvalidSNACErr(snac, w, sequence); err1 != nil {
				err = errors.Join(err1, err)
			}
			if rt.cfg.FailFast {
				panic(err.Error())
			}
			return nil
		}
	}

	return err
}

func (rt *BOSServiceRouter) Signout(ctx context.Context, logger *slog.Logger, sess *user.Session) {
	if err := BroadcastDeparture(ctx, sess, rt.sm, rt.fm); err != nil {
		logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
	}
	rt.sm.Remove(sess)
}

func (rt *BOSServiceRouter) VerifyLogin(conn net.Conn) (*user.Session, uint32, error) {
	seq := uint32(100)

	flap, err := SendAndReceiveSignonFrame(conn, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	ID, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session id from payload")
	}

	sess, ok := rt.sm.Retrieve(string(ID))
	if !ok {
		return nil, 0, fmt.Errorf("unable to find session by id %s", ID)
	}

	return sess, seq, nil
}

func writeOutSNAC(originsnac oscar.SnacFrame, snacFrame oscar.SnacFrame, snacOut any, sequence *uint32, w io.Writer) error {
	if originsnac.RequestID != 0 {
		snacFrame.RequestID = originsnac.RequestID
	}

	snacBuf := &bytes.Buffer{}
	if err := oscar.Marshal(snacFrame, snacBuf); err != nil {
		return err
	}
	if err := oscar.Marshal(snacOut, snacBuf); err != nil {
		return err
	}

	flap := oscar.FlapFrame{
		StartMarker:   42,
		FrameType:     oscar.FlapFrameData,
		Sequence:      uint16(*sequence),
		PayloadLength: uint16(snacBuf.Len()),
	}

	if err := oscar.Marshal(flap, w); err != nil {
		return err
	}

	expectLen := snacBuf.Len()
	c, err := w.Write(snacBuf.Bytes())
	if err != nil {
		return err
	}
	if c != expectLen {
		panic("did not write the expected # of bytes")
	}

	*sequence++
	return nil
}

func sendInvalidSNACErr(snac oscar.SnacFrame, w io.Writer, sequence *uint32) error {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: snac.FoodGroup,
		SubGroup:  0x01, // error subgroup for all SNACs
	}
	snacPayloadOut := oscar.SnacError{
		Code: oscar.ErrorCodeInvalidSnac,
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

type ChatServiceRouter struct {
	ChatRouter
	OServiceChatRouter
	cfg Config
	RouteLogger
}

func (rt *ChatServiceRouter) Route(ctx context.Context, sess *user.Session, r io.Reader, w io.Writer, sequence *uint32, chatSessMgr ChatSessionManager, room ChatRoom) error {
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, r); err != nil {
		return err
	}

	err := func() error {
		switch snac.FoodGroup {
		case oscar.OSERVICE:
			return rt.RouteOService(ctx, sess, chatSessMgr, room, snac, r, w, sequence)
		case oscar.CHAT:
			return rt.RouteChat(ctx, sess, chatSessMgr, snac, r, w, sequence)
		default:
			return ErrUnsupportedSubGroup
		}
	}()

	if err != nil {
		rt.logRequestError(ctx, snac, err)
		if errors.Is(err, ErrUnsupportedSubGroup) {
			if err1 := sendInvalidSNACErr(snac, w, sequence); err1 != nil {
				err = errors.Join(err1, err)
			}
			if rt.cfg.FailFast {
				panic(err.Error())
			}
			return nil
		}
	}

	return err
}
