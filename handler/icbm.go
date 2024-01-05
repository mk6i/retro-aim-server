package handler

import (
	"context"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

const (
	evilDelta     = uint16(100)
	evilDeltaAnon = uint16(30)
)

// NewICBMService returns a new instance of ICBMService.
func NewICBMService(messageRelayer MessageRelayer, feedbagManager FeedbagManager) *ICBMService {
	return &ICBMService{messageRelayer: messageRelayer, feedbagManager: feedbagManager}
}

// ICBMService implements the ICBM food group, which is responsible for sending
// and receiving instant messages and associated functionality such as warning,
// typing events, etc.
type ICBMService struct {
	messageRelayer MessageRelayer
	feedbagManager FeedbagManager
}

// ParameterQueryHandler returns ICBM service parameters.
func (s ICBMService) ParameterQueryHandler(_ context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMParameterReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x04_0x05_ICBMParameterReply{
			MaxSlots:             100,
			ICBMFlags:            3,
			MaxIncomingICBMLen:   512,
			MaxSourceEvil:        999,
			MaxDestinationEvil:   999,
			MinInterICBMInterval: 0,
		},
	}
}

// ChannelMsgToHostHandler relays the instant message SNAC
// oscar.ICBMChannelMsgToHost from the sender to the intended recipient. It
// returns oscar.ICBMHostAck if the oscar.ICBMChannelMsgToHost message contains
// a request acknowledgement flag.
func (s ICBMService) ChannelMsgToHostHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*oscar.SNACMessage, error) {
	blocked, err := s.feedbagManager.BlockedState(sess.ScreenName(), inBody.ScreenName)
	if err != nil {
		return nil, err
	}

	if blocked != state.BlockedNo {
		code := oscar.ErrorCodeNotLoggedOn
		if blocked == state.BlockedA {
			code = oscar.ErrorCodeInLocalPermitDeny
		}
		return &oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: oscar.SNACError{
				Code: code,
			},
		}, nil
	}

	recipSess := s.messageRelayer.RetrieveByScreenName(inBody.ScreenName)
	if recipSess == nil {
		return &oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: oscar.SNACError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	clientIM := oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
		Cookie:    inBody.Cookie,
		ChannelID: inBody.ChannelID,
		TLVUserInfo: oscar.TLVUserInfo{
			ScreenName:   sess.ScreenName(),
			WarningLevel: sess.Warning(),
		},
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					Tag:   0x0B,
					Value: []byte{},
				},
			},
		},
	}
	// copy over TLVs from sender SNAC to recipient SNAC verbatim. this
	// includes ICBMTLVTagRequestHostAck, which is ignored by the client, as
	// far as I can tell.
	clientIM.AppendList(inBody.TLVRestBlock.TLVList)

	s.messageRelayer.RelayToScreenName(ctx, recipSess.ScreenName(), oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMChannelMsgToClient,
		},
		Body: clientIM,
	})

	if _, requestedConfirmation := inBody.TLVRestBlock.Slice(oscar.ICBMTLVTagRequestHostAck); !requestedConfirmation {
		// don't ack message
		return nil, nil
	}

	// ack message back to sender
	return &oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMHostAck,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x04_0x0C_ICBMHostAck{
			Cookie:     inBody.Cookie,
			ChannelID:  inBody.ChannelID,
			ScreenName: inBody.ScreenName,
		},
	}, nil
}

// ClientEventHandler relays SNAC oscar.ICBMClientEvent typing events from the
// sender to the recipient.
func (s ICBMService) ClientEventHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x04_0x14_ICBMClientEvent) error {
	blocked, err := s.feedbagManager.BlockedState(sess.ScreenName(), inBody.ScreenName)

	switch {
	case err != nil:
		return err
	case blocked != state.BlockedNo:
		return nil
	default:
		s.messageRelayer.RelayToScreenName(ctx, inBody.ScreenName, oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMClientEvent,
				RequestID: inFrame.RequestID,
			},
			Body: oscar.SNAC_0x04_0x14_ICBMClientEvent{
				Cookie:     inBody.Cookie,
				ChannelID:  inBody.ChannelID,
				ScreenName: sess.ScreenName(),
				Event:      inBody.Event,
			},
		})
		return nil
	}
}

// EvilRequestHandler handles user warning (a.k.a evil) notifications. It
// receives oscar.ICBMEvilRequest warning SNAC, increments the warned user's
// warning level, and sends the warned user a notification informing them that
// they have been warned. The user may choose to warn anonymously or
// non-anonymously. It returns SNAC oscar.ICBMEvilReply to confirm that the
// warning was sent. Users may not warn themselves or warn users they have
// blocked or are blocked by.
func (s ICBMService) EvilRequestHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x04_0x08_ICBMEvilRequest) (oscar.SNACMessage, error) {
	// don't let users warn themselves, it causes the AIM client to go into a
	// weird state.
	if inBody.ScreenName == sess.ScreenName() {
		return oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: oscar.SNACError{
				Code: oscar.ErrorCodeNotSupportedByHost,
			},
		}, nil
	}

	blocked, err := s.feedbagManager.BlockedState(sess.ScreenName(), inBody.ScreenName)
	if err != nil {
		return oscar.SNACMessage{}, nil
	}
	if blocked != state.BlockedNo {
		return oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.ICBM,
				SubGroup:  oscar.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: oscar.SNACError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	recipSess := s.messageRelayer.RetrieveByScreenName(inBody.ScreenName)
	if recipSess == nil {
		return oscar.SNACMessage{}, nil
	}

	increase := evilDelta
	if inBody.SendAs == 1 {
		increase = evilDeltaAnon
	}
	recipSess.IncrementWarning(increase)

	var notif any
	if inBody.SendAs == 0 {
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

	s.messageRelayer.RelayToScreenName(ctx, recipSess.ScreenName(), oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceEvilNotification,
		},
		Body: notif,
	})

	// inform the warned user's buddies that their warning level has increased
	if err := broadcastArrival(ctx, recipSess, s.messageRelayer, s.feedbagManager); err != nil {
		return oscar.SNACMessage{}, nil
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ICBM,
			SubGroup:  oscar.ICBMEvilReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x04_0x09_ICBMEvilReply{
			EvilDeltaApplied: increase,
			UpdatedEvilValue: recipSess.Warning(),
		},
	}, nil
}
