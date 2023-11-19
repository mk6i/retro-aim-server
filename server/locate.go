package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
)

type LocateHandler interface {
	RightsQueryHandler(ctx context.Context) oscar.XMessage
	SetDirInfoHandler(ctx context.Context) oscar.XMessage
	SetInfoHandler(ctx context.Context, sess *Session, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error
	SetKeywordInfoHandler(ctx context.Context) oscar.XMessage
	UserInfoQuery2Handler(ctx context.Context, sess *Session, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (oscar.XMessage, error)
}

func NewLocateRouter(handler LocateHandler, logger *slog.Logger) LocateRouter {
	return LocateRouter{
		LocateHandler: handler,
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type LocateRouter struct {
	LocateHandler
	RouteLogger
}

func (rt LocateRouter) RouteLocate(ctx context.Context, sess *Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.LocateRightsQuery:
		outSNAC := rt.RightsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.LocateSetInfo:
		inSNAC := oscar.SNAC_0x02_0x04_LocateSetInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.SetInfoHandler(ctx, sess, inSNAC)
	case oscar.LocateSetDirInfo:
		inSNAC := oscar.SNAC_0x02_0x09_LocateSetDirInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.SetDirInfoHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.LocateGetDirInfo:
		inSNAC := oscar.SNAC_0x02_0x0B_LocateGetDirInfo{}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return oscar.Unmarshal(&inSNAC, r)
	case oscar.LocateSetKeywordInfo:
		inSNAC := oscar.SNAC_0x02_0x0F_LocateSetKeywordInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.SetKeywordInfoHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.LocateUserInfoQuery2:
		inSNAC := oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.UserInfoQuery2Handler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}
