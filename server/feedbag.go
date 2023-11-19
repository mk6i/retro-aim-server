package server

import (
	"context"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"log/slog"
)

type FeedbagHandler interface {
	DeleteItemHandler(ctx context.Context, sess *Session, snacPayloadIn oscar.SNAC_0x13_0x0A_FeedbagDeleteItem) (oscar.XMessage, error)
	InsertItemHandler(ctx context.Context, sess *Session, snacPayloadIn oscar.SNAC_0x13_0x08_FeedbagInsertItem) (oscar.XMessage, error)
	QueryHandler(ctx context.Context, sess *Session) (oscar.XMessage, error)
	QueryIfModifiedHandler(ctx context.Context, sess *Session, snacPayloadIn oscar.SNAC_0x13_0x05_FeedbagQueryIfModified) (oscar.XMessage, error)
	RightsQueryHandler(context.Context) oscar.XMessage
	StartClusterHandler(context.Context, oscar.SNAC_0x13_0x11_FeedbagStartCluster)
	UpdateItemHandler(ctx context.Context, sess *Session, snacPayloadIn oscar.SNAC_0x13_0x09_FeedbagUpdateItem) (oscar.XMessage, error)
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

func (rt FeedbagRouter) RouteFeedbag(ctx context.Context, sess *Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.FeedbagRightsQuery:
		inSNAC := oscar.SNAC_0x13_0x02_FeedbagRightsQuery{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.RightsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagQuery:
		inSNAC, err := rt.QueryHandler(ctx, sess)
		if err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return writeOutSNAC(SNACFrame, inSNAC.SnacFrame, inSNAC.SnacOut, sequence, w)
	case oscar.FeedbagQueryIfModified:
		inSNAC := oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.QueryIfModifiedHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagUse:
		rt.logRequest(ctx, SNACFrame, nil)
		return nil
	case oscar.FeedbagInsertItem:
		inSNAC := oscar.SNAC_0x13_0x08_FeedbagInsertItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.InsertItemHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagUpdateItem:
		inSNAC := oscar.SNAC_0x13_0x09_FeedbagUpdateItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.UpdateItemHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagDeleteItem:
		inSNAC := oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.DeleteItemHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagStartCluster:
		inSNAC := oscar.SNAC_0x13_0x11_FeedbagStartCluster{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.StartClusterHandler(ctx, inSNAC)
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return nil
	case oscar.FeedbagEndCluster:
		rt.logRequest(ctx, SNACFrame, nil)
		return nil
	default:
		return ErrUnsupportedSubGroup
	}
}
