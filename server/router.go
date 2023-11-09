package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"log/slog"
	"net"
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
	}
}

func NewChatServiceRouter(logger *slog.Logger, cfg Config, fm FeedbagManager, sm SessionManager) ChatServiceRouter {
	return ChatServiceRouter{
		OServiceChatRouter: NewOServiceRouterForChat(logger, cfg, fm, sm),
		ChatRouter:         NewChatRouter(logger),
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
	sm SessionManager
	fm FeedbagManager
}

func (rt *BOSServiceRouter) Route(ctx context.Context, sess *Session, w io.Writer, sequence *uint32, snac oscar.SnacFrame, buf io.Reader) error {
	switch snac.FoodGroup {
	case oscar.OSERVICE:
		return rt.RouteOService(ctx, sess, snac, buf, w, sequence)
	case oscar.LOCATE:
		return rt.RouteLocate(ctx, sess, snac, buf, w, sequence)
	case oscar.BUDDY:
		return rt.RouteBuddy(ctx, snac, buf, w, sequence)
	case oscar.ICBM:
		return rt.RouteICBM(ctx, sess, snac, buf, w, sequence)
	case oscar.CHAT_NAV:
		return rt.RouteChatNav(ctx, sess, snac, buf, w, sequence)
	case oscar.FEEDBAG:
		return rt.RouteFeedbag(ctx, sess, snac, buf, w, sequence)
	case oscar.BUCP:
		return routeBUCP(ctx)
	case oscar.ALERT:
		return rt.RouteAlert(ctx, snac)
	default:
		return ErrUnsupportedFoodGroup
	}
}

func (rt *BOSServiceRouter) Signout(ctx context.Context, logger *slog.Logger, sess *Session) {
	if err := BroadcastDeparture(ctx, sess, rt.sm, rt.fm); err != nil {
		logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
	}
	rt.sm.Remove(sess)
}

func (rt *BOSServiceRouter) VerifyLogin(conn net.Conn) (*Session, uint32, error) {
	seq := uint32(100)

	flap, err := SendAndReceiveSignonFrame(conn, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	ID, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session ID from payload")
	}

	sess, ok := rt.sm.Retrieve(string(ID))
	if !ok {
		return nil, 0, fmt.Errorf("unable to find session by ID %s", ID)
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
}

func (rt *ChatServiceRouter) Route(ctx context.Context, sess *Session, w io.Writer, sequence *uint32, snac oscar.SnacFrame, buf io.Reader, room ChatRoom) error {
	switch snac.FoodGroup {
	case oscar.OSERVICE:
		return rt.RouteOService(ctx, sess, room, snac, buf, w, sequence)
	case oscar.CHAT:
		return rt.RouteChat(ctx, sess, room, snac, buf, w, sequence)
	default:
		return ErrUnsupportedFoodGroup
	}
}
