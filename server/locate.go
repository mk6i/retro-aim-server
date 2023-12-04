package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type LocateHandler interface {
	RightsQueryHandler(ctx context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage
	SetDirInfoHandler(ctx context.Context, frame oscar.SNACFrame) oscar.SNACMessage
	SetInfoHandler(ctx context.Context, sess *state.Session, inBody oscar.SNAC_0x02_0x04_LocateSetInfo) error
	SetKeywordInfoHandler(ctx context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage
	UserInfoQuery2Handler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (oscar.SNACMessage, error)
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

func (rt LocateRouter) RouteLocate(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.LocateRightsQuery:
		outSNAC := rt.RightsQueryHandler(ctx, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.LocateSetInfo:
		inBody := oscar.SNAC_0x02_0x04_LocateSetInfo{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		rt.logRequest(ctx, inFrame, inBody)
		return rt.SetInfoHandler(ctx, sess, inBody)
	case oscar.LocateSetDirInfo:
		inBody := oscar.SNAC_0x02_0x09_LocateSetDirInfo{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC := rt.SetDirInfoHandler(ctx, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.LocateGetDirInfo:
		inBody := oscar.SNAC_0x02_0x0B_LocateGetDirInfo{}
		rt.logRequest(ctx, inFrame, inBody)
		return oscar.Unmarshal(&inBody, r)
	case oscar.LocateSetKeywordInfo:
		inBody := oscar.SNAC_0x02_0x0F_LocateSetKeywordInfo{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC := rt.SetKeywordInfoHandler(ctx, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.LocateUserInfoQuery2:
		inBody := oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.UserInfoQuery2Handler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}
