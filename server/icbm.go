package server

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
)

const (
	evilDelta     = uint16(100)
	evilDeltaAnon = uint16(30)
)

type ICBMHandler interface {
	ChannelMsgToHostHandler(ctx context.Context, sm SessionManager, fm FeedbagManager, sess *Session, snacPayloadIn oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*XMessage, error)
	ClientEventHandler(ctx context.Context, sm SessionManager, fm FeedbagManager, sess *Session, snacPayloadIn oscar.SNAC_0x04_0x14_ICBMClientEvent) error
	EvilRequestHandler(ctx context.Context, sm SessionManager, fm FeedbagManager, sess *Session, snacPayloadIn oscar.SNAC_0x04_0x08_ICBMEvilRequest) (XMessage, error)
	ParameterQueryHandler(context.Context) XMessage
}

func NewICBMRouter(logger *slog.Logger) ICBMRouter {
	return ICBMRouter{
		ICBMHandler: ICBMService{},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type ICBMRouter struct {
	ICBMHandler
	RouteLogger
}

func (rt *ICBMRouter) RouteICBM(ctx context.Context, sm SessionManager, fm FeedbagManager, sess *Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.ICBMAddParameters:
		inSNAC := oscar.SNAC_0x04_0x02_ICBMAddParameters{}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return oscar.Unmarshal(&inSNAC, r)
	case oscar.ICBMParameterQuery:
		outSNAC := rt.ParameterQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, outSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.ICBMChannelMsgToHost:
		inSNAC := oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ChannelMsgToHostHandler(ctx, sm, fm, sess, inSNAC)
		if err != nil || outSNAC == nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user sent an IM", slog.String("recipient", inSNAC.ScreenName))
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.ICBMEvilRequest:
		inSNAC := oscar.SNAC_0x04_0x08_ICBMEvilRequest{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.EvilRequestHandler(ctx, sm, fm, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.ICBMClientErr:
		inSNAC := oscar.SNAC_0x04_0x0B_ICBMClientErr{}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return oscar.Unmarshal(&inSNAC, r)
	case oscar.ICBMClientEvent:
		inSNAC := oscar.SNAC_0x04_0x14_ICBMClientEvent{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.ClientEventHandler(ctx, sm, fm, sess, inSNAC)
	default:
		return ErrUnsupportedSubGroup
	}
}

type ICBMService struct {
}

func (s ICBMService) ParameterQueryHandler(context.Context) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMParameterReply,
		},
		snacOut: oscar.SNAC_0x04_0x05_ICBMParameterReply{
			MaxSlots:             100,
			ICBMFlags:            3,
			MaxIncomingICBMLen:   512,
			MaxSourceEvil:        999,
			MaxDestinationEvil:   999,
			MinInterICBMInterval: 0,
		},
	}
}

func (s ICBMService) ChannelMsgToHostHandler(ctx context.Context, sm SessionManager, fm FeedbagManager, sess *Session, snacPayloadIn oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*XMessage, error) {
	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)
	if err != nil {
		return nil, err
	}

	if blocked != BlockedNo {
		code := oscar.ErrorCodeNotLoggedOn
		if blocked == BlockedA {
			code = oscar.ErrorCodeInLocalPermitDeny
		}
		return &XMessage{
			snacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
			},
			snacOut: oscar.SnacError{
				Code: code,
			},
		}, nil
	}

	recipSess, err := sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	switch {
	case errors.Is(err, ErrSessNotFound):
		return &XMessage{
			snacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
			},
			snacOut: oscar.SnacError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	case err != nil:
		return nil, err
	}

	clientIM := oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
		Cookie:    snacPayloadIn.Cookie,
		ChannelID: snacPayloadIn.ChannelID,
		TLVUserInfo: oscar.TLVUserInfo{
			ScreenName:   sess.ScreenName,
			WarningLevel: sess.GetWarning(),
		},
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: 0x0B,
					Val:   []byte{},
				},
			},
		},
	}
	// copy over TLVs from sender SNAC to recipient SNAC verbatim. this
	// includes ICBMTLVTagRequestHostAck, which is ignored by the client, as
	// far as I can tell.
	clientIM.AddTLVList(snacPayloadIn.TLVRestBlock.TLVList)

	sm.SendToScreenName(ctx, recipSess.ScreenName, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMChannelMsgToclient,
		},
		snacOut: clientIM,
	})

	if _, requestedConfirmation := snacPayloadIn.TLVRestBlock.GetSlice(oscar.ICBMTLVTagRequestHostAck); !requestedConfirmation {
		// don't ack message
		return nil, nil
	}

	// ack message back to sender
	return &XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMHostAck,
		},
		snacOut: oscar.SNAC_0x04_0x0C_ICBMHostAck{
			Cookie:     snacPayloadIn.Cookie,
			ChannelID:  snacPayloadIn.ChannelID,
			ScreenName: snacPayloadIn.ScreenName,
		},
	}, nil
}

func (s ICBMService) ClientEventHandler(ctx context.Context, sm SessionManager, fm FeedbagManager, sess *Session, snacPayloadIn oscar.SNAC_0x04_0x14_ICBMClientEvent) error {
	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)

	switch {
	case err != nil:
		return err
	case blocked != BlockedNo:
		return nil
	default:
		sm.SendToScreenName(ctx, snacPayloadIn.ScreenName, XMessage{
			snacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMClientEvent,
			},
			snacOut: oscar.SNAC_0x04_0x14_ICBMClientEvent{
				Cookie:     snacPayloadIn.Cookie,
				ChannelID:  snacPayloadIn.ChannelID,
				ScreenName: sess.ScreenName,
				Event:      snacPayloadIn.Event,
			},
		})
		return nil
	}
}

func (s ICBMService) EvilRequestHandler(ctx context.Context, sm SessionManager, fm FeedbagManager, sess *Session, snacPayloadIn oscar.SNAC_0x04_0x08_ICBMEvilRequest) (XMessage, error) {
	// don't let users warn themselves, it causes the AIM client to go into a
	// weird state.
	if snacPayloadIn.ScreenName == sess.ScreenName {
		return XMessage{
			snacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
			},
			snacOut: oscar.SnacError{
				Code: oscar.ErrorCodeNotSupportedByHost,
			},
		}, nil
	}

	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)
	if err != nil {
		return XMessage{}, nil
	}
	if blocked != BlockedNo {
		return XMessage{
			snacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
			},
			snacOut: oscar.SnacError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	recipSess, err := sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	if err != nil {
		return XMessage{}, nil
	}

	increase := evilDelta
	if snacPayloadIn.SendAs == 1 {
		increase = evilDeltaAnon
	}
	recipSess.IncreaseWarning(increase)

	var notif any
	if snacPayloadIn.SendAs == 0 {
		notif = oscar.SNAC_0x01_0x10_OServiceEvilNotification{
			NewEvil: recipSess.GetWarning(),
			TLVUserInfo: oscar.TLVUserInfo{
				ScreenName:   sess.ScreenName,
				WarningLevel: recipSess.GetWarning(),
			},
		}
	} else {
		notif = oscar.SNAC_0x01_0x10_OServiceEvilNotificationAnon{
			NewEvil: recipSess.GetWarning(),
		}
	}

	sm.SendToScreenName(ctx, recipSess.ScreenName, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceEvilNotification,
		},
		snacOut: notif,
	})

	if err := BroadcastArrival(ctx, recipSess, sm, fm); err != nil {
		return XMessage{}, nil
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMEvilReply,
		},
		snacOut: oscar.SNAC_0x04_0x09_ICBMEvilReply{
			EvilDeltaApplied: increase,
			UpdatedEvilValue: recipSess.GetWarning(),
		},
	}, nil
}
