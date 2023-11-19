package handler

import (
	"context"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
)

const (
	evilDelta     = uint16(100)
	evilDeltaAnon = uint16(30)
)

func NewICBMService(sm server.SessionManager, fm server.FeedbagManager) *ICBMService {
	return &ICBMService{sm: sm, fm: fm}
}

type ICBMService struct {
	sm server.SessionManager
	fm server.FeedbagManager
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

func (s ICBMService) ChannelMsgToHostHandler(ctx context.Context, sess *server.Session, snacPayloadIn oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*oscar.XMessage, error) {
	blocked, err := s.fm.Blocked(sess.ScreenName(), snacPayloadIn.ScreenName)
	if err != nil {
		return nil, err
	}

	if blocked != server.BlockedNo {
		code := oscar.ErrorCodeNotLoggedOn
		if blocked == server.BlockedA {
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

func (s ICBMService) ClientEventHandler(ctx context.Context, sess *server.Session, snacPayloadIn oscar.SNAC_0x04_0x14_ICBMClientEvent) error {
	blocked, err := s.fm.Blocked(sess.ScreenName(), snacPayloadIn.ScreenName)

	switch {
	case err != nil:
		return err
	case blocked != server.BlockedNo:
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

func (s ICBMService) EvilRequestHandler(ctx context.Context, sess *server.Session, snacPayloadIn oscar.SNAC_0x04_0x08_ICBMEvilRequest) (oscar.XMessage, error) {
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
	if blocked != server.BlockedNo {
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

	if err := broadcastArrival(ctx, recipSess, s.sm, s.fm); err != nil {
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
