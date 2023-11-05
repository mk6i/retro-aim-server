package server

import (
	"bytes"
	"context"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"log/slog"
)

func NewRouter(logger *slog.Logger) Router {
	return Router{
		AlertRouter:    NewAlertRouter(logger),
		BuddyRouter:    NewBuddyRouter(logger),
		ChatNavRouter:  NewChatNavRouter(logger),
		ChatRouter:     NewChatRouter(logger),
		FeedbagRouter:  NewFeedbagRouter(logger),
		ICBMRouter:     NewICBMRouter(logger),
		LocateRouter:   NewLocateRouter(logger),
		OServiceRouter: NewOServiceRouter(logger),
	}
}

func NewRouterForChat(logger *slog.Logger) Router {
	r := NewRouter(logger)
	r.OServiceRouter = NewOServiceRouterForChat(logger)
	return r
}

type Router struct {
	AlertRouter
	BuddyRouter
	ChatNavRouter
	ChatRouter
	FeedbagRouter
	ICBMRouter
	LocateRouter
	OServiceRouter
}

func (rt *Router) routeIncomingRequests(ctx context.Context, cfg Config, sm SessionManager, sess *Session, fm *FeedbagStore, cr *ChatRegistry, rw io.ReadWriter, sequence *uint32, snac oscar.SnacFrame, buf io.Reader, room ChatRoom) error {
	switch snac.FoodGroup {
	case oscar.OSERVICE:
		return rt.RouteOService(ctx, cfg, cr, sm, fm, sess, room, snac, buf, rw, sequence)
	case oscar.LOCATE:
		return rt.RouteLocate(ctx, sess, sm, fm, snac, buf, rw, sequence)
	case oscar.BUDDY:
		return rt.RouteBuddy(ctx, snac, buf, rw, sequence)
	case oscar.ICBM:
		return rt.RouteICBM(ctx, sm, fm, sess, snac, buf, rw, sequence)
	case oscar.CHAT_NAV:
		return rt.RouteChatNav(ctx, sess, cr, snac, buf, rw, sequence)
	case oscar.FEEDBAG:
		return rt.RouteFeedbag(ctx, sm, sess, fm, snac, buf, rw, sequence)
	case oscar.BUCP:
		return routeBUCP(ctx)
	case oscar.CHAT:
		return rt.RouteChat(ctx, sess, sm, snac, buf, rw, sequence)
	case oscar.ALERT:
		return rt.RouteAlert(ctx, snac)
	default:
		return ErrUnsupportedFoodGroup
	}
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
