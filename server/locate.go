package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type LocateHandler interface {
	RightsQueryHandler(ctx context.Context) oscar.SNACMessage
	SetDirInfoHandler(ctx context.Context) oscar.SNACMessage
	SetInfoHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error
	SetKeywordInfoHandler(ctx context.Context) oscar.SNACMessage
	UserInfoQuery2Handler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (oscar.SNACMessage, error)
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

func (rt LocateRouter) RouteLocate(ctx context.Context, sess *state.Session, SNACFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.LocateRightsQuery:
		outSNAC := rt.RightsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame, outSNAC.Frame, outSNAC.Body, sequence, w)
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
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame, outSNAC.Frame, outSNAC.Body, sequence, w)
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
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame, outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.LocateUserInfoQuery2:
		inSNAC := oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.UserInfoQuery2Handler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame, outSNAC.Frame, outSNAC.Body, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}
