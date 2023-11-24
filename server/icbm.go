package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type ICBMHandler interface {
	ChannelMsgToHostHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*oscar.SNACMessage, error)
	ClientEventHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x04_0x14_ICBMClientEvent) error
	EvilRequestHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x04_0x08_ICBMEvilRequest) (oscar.SNACMessage, error)
	ParameterQueryHandler(ctx context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage
}

func NewICBMRouter(logger *slog.Logger, handler ICBMHandler) ICBMRouter {
	return ICBMRouter{
		ICBMHandler: handler,
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type ICBMRouter struct {
	ICBMHandler
	RouteLogger
}

func (rt *ICBMRouter) RouteICBM(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.ICBMAddParameters:
		inBody := oscar.SNAC_0x04_0x02_ICBMAddParameters{}
		rt.logRequest(ctx, inFrame, inBody)
		return oscar.Unmarshal(&inBody, r)
	case oscar.ICBMParameterQuery:
		outSNAC := rt.ParameterQueryHandler(ctx, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, outSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.ICBMChannelMsgToHost:
		inBody := oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.ChannelMsgToHostHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user sent an IM", slog.String("recipient", inBody.ScreenName))
		if outSNAC == nil {
			return nil
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.ICBMEvilRequest:
		inBody := oscar.SNAC_0x04_0x08_ICBMEvilRequest{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.EvilRequestHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.ICBMClientErr:
		inBody := oscar.SNAC_0x04_0x0B_ICBMClientErr{}
		rt.logRequest(ctx, inFrame, inBody)
		return oscar.Unmarshal(&inBody, r)
	case oscar.ICBMClientEvent:
		inBody := oscar.SNAC_0x04_0x14_ICBMClientEvent{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		rt.logRequest(ctx, inFrame, inBody)
		return rt.ClientEventHandler(ctx, sess, inFrame, inBody)
	default:
		return ErrUnsupportedSubGroup
	}
}
