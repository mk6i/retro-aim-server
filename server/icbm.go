package server

import (
	"context"
	"github.com/mkaminski/goaim/user"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
)

const (
	evilDelta     = uint16(100)
	evilDeltaAnon = uint16(30)
)

type ICBMHandler interface {
	ChannelMsgToHostHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*oscar.XMessage, error)
	ClientEventHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x04_0x14_ICBMClientEvent) error
	EvilRequestHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x04_0x08_ICBMEvilRequest) (oscar.XMessage, error)
	ParameterQueryHandler(context.Context) oscar.XMessage
}

func NewICBMRouter(logger *slog.Logger, sm SessionManager, fm FeedbagManager) ICBMRouter {
	return ICBMRouter{
		ICBMHandler: ICBMService{
			sm: sm,
			fm: fm,
		},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type ICBMRouter struct {
	ICBMHandler
	RouteLogger
}

func (rt *ICBMRouter) RouteICBM(ctx context.Context, sess *user.Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.ICBMAddParameters:
		inSNAC := oscar.SNAC_0x04_0x02_ICBMAddParameters{}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return oscar.Unmarshal(&inSNAC, r)
	case oscar.ICBMParameterQuery:
		outSNAC := rt.ParameterQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, outSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.ICBMChannelMsgToHost:
		inSNAC := oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ChannelMsgToHostHandler(ctx, sess, inSNAC)
		if err != nil || outSNAC == nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user sent an IM", slog.String("recipient", inSNAC.ScreenName))
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.ICBMEvilRequest:
		inSNAC := oscar.SNAC_0x04_0x08_ICBMEvilRequest{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.EvilRequestHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
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
		return rt.ClientEventHandler(ctx, sess, inSNAC)
	default:
		return ErrUnsupportedSubGroup
	}
}

type ICBMService struct {
	sm SessionManager
	fm FeedbagManager
}

func (s ICBMService) ParameterQueryHandler(context.Context) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMParameterReply,
		},
		SnacOut: oscar.SNAC_0x04_0x05_ICBMParameterReply{
			MaxSlots:             100,
			ICBMFlags:            3,
			MaxIncomingICBMLen:   512,
			MaxSourceEvil:        999,
			MaxDestinationEvil:   999,
			MinInterICBMInterval: 0,
		},
	}
}

func (s ICBMService) ChannelMsgToHostHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*oscar.XMessage, error) {
	blocked, err := s.fm.Blocked(sess.ScreenName(), snacPayloadIn.ScreenName)
	if err != nil {
		return nil, err
	}

	if blocked != user.BlockedNo {
		code := oscar.ErrorCodeNotLoggedOn
		if blocked == user.BlockedA {
			code = oscar.ErrorCodeInLocalPermitDeny
		}
		return &oscar.XMessage{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
			},
			SnacOut: oscar.SnacError{
				Code: code,
			},
		}, nil
	}

	recipSess := s.sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	if recipSess == nil {
		return &oscar.XMessage{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
			},
			SnacOut: oscar.SnacError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	clientIM := oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
		Cookie:    snacPayloadIn.Cookie,
		ChannelID: snacPayloadIn.ChannelID,
		TLVUserInfo: oscar.TLVUserInfo{
			ScreenName:   sess.ScreenName(),
			WarningLevel: sess.Warning(),
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

	s.sm.SendToScreenName(ctx, recipSess.ScreenName(), oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMChannelMsgToclient,
		},
		SnacOut: clientIM,
	})

	if _, requestedConfirmation := snacPayloadIn.TLVRestBlock.GetSlice(oscar.ICBMTLVTagRequestHostAck); !requestedConfirmation {
		// don't ack message
		return nil, nil
	}

	// ack message back to sender
	return &oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMHostAck,
		},
		SnacOut: oscar.SNAC_0x04_0x0C_ICBMHostAck{
			Cookie:     snacPayloadIn.Cookie,
			ChannelID:  snacPayloadIn.ChannelID,
			ScreenName: snacPayloadIn.ScreenName,
		},
	}, nil
}

func (s ICBMService) ClientEventHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x04_0x14_ICBMClientEvent) error {
	blocked, err := s.fm.Blocked(sess.ScreenName(), snacPayloadIn.ScreenName)

	switch {
	case err != nil:
		return err
	case blocked != user.BlockedNo:
		return nil
	default:
		s.sm.SendToScreenName(ctx, snacPayloadIn.ScreenName, oscar.XMessage{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMClientEvent,
			},
			SnacOut: oscar.SNAC_0x04_0x14_ICBMClientEvent{
				Cookie:     snacPayloadIn.Cookie,
				ChannelID:  snacPayloadIn.ChannelID,
				ScreenName: sess.ScreenName(),
				Event:      snacPayloadIn.Event,
			},
		})
		return nil
	}
}

func (s ICBMService) EvilRequestHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x04_0x08_ICBMEvilRequest) (oscar.XMessage, error) {
	// don't let users warn themselves, it causes the AIM client to go into a
	// weird state.
	if snacPayloadIn.ScreenName == sess.ScreenName() {
		return oscar.XMessage{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
			},
			SnacOut: oscar.SnacError{
				Code: oscar.ErrorCodeNotSupportedByHost,
			},
		}, nil
	}

	blocked, err := s.fm.Blocked(sess.ScreenName(), snacPayloadIn.ScreenName)
	if err != nil {
		return oscar.XMessage{}, nil
	}
	if blocked != user.BlockedNo {
		return oscar.XMessage{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
			},
			SnacOut: oscar.SnacError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	recipSess := s.sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	if recipSess == nil {
		return oscar.XMessage{}, nil
	}

	increase := evilDelta
	if snacPayloadIn.SendAs == 1 {
		increase = evilDeltaAnon
	}
	recipSess.IncreaseWarning(increase)

	var notif any
	if snacPayloadIn.SendAs == 0 {
		notif = oscar.SNAC_0x01_0x10_OServiceEvilNotification{
			NewEvil: recipSess.Warning(),
			TLVUserInfo: oscar.TLVUserInfo{
				ScreenName:   sess.ScreenName(),
				WarningLevel: recipSess.Warning(),
			},
		}
	} else {
		notif = oscar.SNAC_0x01_0x10_OServiceEvilNotificationAnon{
			NewEvil: recipSess.Warning(),
		}
	}

	s.sm.SendToScreenName(ctx, recipSess.ScreenName(), oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceEvilNotification,
		},
		SnacOut: notif,
	})

	if err := BroadcastArrival(ctx, recipSess, s.sm, s.fm); err != nil {
		return oscar.XMessage{}, nil
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMEvilReply,
		},
		SnacOut: oscar.SNAC_0x04_0x09_ICBMEvilReply{
			EvilDeltaApplied: increase,
			UpdatedEvilValue: recipSess.Warning(),
		},
	}, nil
}
