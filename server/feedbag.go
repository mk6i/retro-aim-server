package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type FeedbagHandler interface {
	DeleteItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x0A_FeedbagDeleteItem) (oscar.SNACMessage, error)
	InsertItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x08_FeedbagInsertItem) (oscar.SNACMessage, error)
	QueryHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame) (oscar.SNACMessage, error)
	QueryIfModifiedHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x05_FeedbagQueryIfModified) (oscar.SNACMessage, error)
	RightsQueryHandler(ctx context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage
	StartClusterHandler(ctx context.Context, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x11_FeedbagStartCluster)
	UpdateItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x09_FeedbagUpdateItem) (oscar.SNACMessage, error)
}

func NewFeedbagRouter(logger *slog.Logger, handler FeedbagHandler) FeedbagRouter {
	return FeedbagRouter{
		FeedbagHandler: handler,
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type FeedbagRouter struct {
	FeedbagHandler
	RouteLogger
}

func (rt FeedbagRouter) RouteFeedbag(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.FeedbagRightsQuery:
		inBody := oscar.SNAC_0x13_0x02_FeedbagRightsQuery{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC := rt.RightsQueryHandler(ctx, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.FeedbagQuery:
		outSNAC, err := rt.QueryHandler(ctx, sess, inFrame)
		if err != nil {
			return err
		}
		rt.logRequest(ctx, inFrame, outSNAC)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.FeedbagQueryIfModified:
		inBody := oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.QueryIfModifiedHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.FeedbagUse:
		rt.logRequest(ctx, inFrame, nil)
		return nil
	case oscar.FeedbagInsertItem:
		inBody := oscar.SNAC_0x13_0x08_FeedbagInsertItem{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.InsertItemHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.FeedbagUpdateItem:
		inBody := oscar.SNAC_0x13_0x09_FeedbagUpdateItem{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.UpdateItemHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.FeedbagDeleteItem:
		inBody := oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.DeleteItemHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.FeedbagStartCluster:
		inBody := oscar.SNAC_0x13_0x11_FeedbagStartCluster{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		rt.StartClusterHandler(ctx, inFrame, inBody)
		rt.logRequest(ctx, inFrame, inBody)
		return nil
	case oscar.FeedbagEndCluster:
		rt.logRequest(ctx, inFrame, nil)
		return nil
	default:
		return ErrUnsupportedSubGroup
	}
}
