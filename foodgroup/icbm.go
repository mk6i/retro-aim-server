package foodgroup

import (
	"context"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

const (
	evilDelta     = uint16(100)
	evilDeltaAnon = uint16(30)
)

// NewICBMService returns a new instance of ICBMService.
func NewICBMService(
	messageRelayer MessageRelayer,
	feedbagManager FeedbagManager,
	buddyUpdateBroadcaster BuddyBroadcaster,
) *ICBMService {
	return &ICBMService{
		buddyUpdateBroadcaster: buddyUpdateBroadcaster,
		feedbagManager:         feedbagManager,
		messageRelayer:         messageRelayer,
	}
}

// ICBMService provides functionality for the ICBM food group, which is
// responsible for sending and receiving instant messages and associated
// functionality such as warning, typing events, etc.
type ICBMService struct {
	buddyUpdateBroadcaster BuddyBroadcaster
	feedbagManager         FeedbagManager
	messageRelayer         MessageRelayer
}

// ParameterQuery returns ICBM service parameters.
func (s ICBMService) ParameterQuery(_ context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMParameterReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x04_0x05_ICBMParameterReply{
			MaxSlots:             100,
			ICBMFlags:            3,
			MaxIncomingICBMLen:   512,
			MaxSourceEvil:        999,
			MaxDestinationEvil:   999,
			MinInterICBMInterval: 0,
		},
	}
}

// ChannelMsgToHost relays the instant message SNAC wire.ICBMChannelMsgToHost
// from the sender to the intended recipient. It returns wire.ICBMHostAck if
// the wire.ICBMChannelMsgToHost message contains a request acknowledgement
// flag.
func (s ICBMService) ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*wire.SNACMessage, error) {
	blocked, err := s.feedbagManager.BlockedState(sess.ScreenName(), inBody.ScreenName)
	if err != nil {
		return nil, err
	}

	if blocked != state.BlockedNo {
		code := wire.ErrorCodeNotLoggedOn
		if blocked == state.BlockedA {
			code = wire.ErrorCodeInLocalPermitDeny
		}
		return &wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.ICBM,
				SubGroup:  wire.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: code,
			},
		}, nil
	}

	recipSess := s.messageRelayer.RetrieveByScreenName(inBody.ScreenName)
	if recipSess == nil {
		return &wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.ICBM,
				SubGroup:  wire.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: wire.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	clientIM := wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
		Cookie:    inBody.Cookie,
		ChannelID: inBody.ChannelID,
		TLVUserInfo: wire.TLVUserInfo{
			ScreenName:   sess.ScreenName(),
			WarningLevel: sess.Warning(),
		},
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				{
					Tag:   0x0B,
					Value: []byte{},
				},
			},
		},
	}
	// copy over TLVs from sender SNAC to recipient SNAC verbatim. this
	// includes ICBMTLVRequestHostAck, which is ignored by the client, as
	// far as I can tell.
	clientIM.AppendList(inBody.TLVRestBlock.TLVList)

	s.messageRelayer.RelayToScreenName(ctx, recipSess.ScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMChannelMsgToClient,
		},
		Body: clientIM,
	})

	if _, requestedConfirmation := inBody.TLVRestBlock.Slice(wire.ICBMTLVRequestHostAck); !requestedConfirmation {
		// don't ack message
		return nil, nil
	}

	// ack message back to sender
	return &wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMHostAck,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x04_0x0C_ICBMHostAck{
			Cookie:     inBody.Cookie,
			ChannelID:  inBody.ChannelID,
			ScreenName: inBody.ScreenName,
		},
	}, nil
}

// ClientEvent relays SNAC wire.ICBMClientEvent typing events from the
// sender to the recipient.
func (s ICBMService) ClientEvent(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x14_ICBMClientEvent) error {
	blocked, err := s.feedbagManager.BlockedState(sess.ScreenName(), inBody.ScreenName)

	switch {
	case err != nil:
		return err
	case blocked != state.BlockedNo:
		return nil
	default:
		s.messageRelayer.RelayToScreenName(ctx, inBody.ScreenName, wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.ICBM,
				SubGroup:  wire.ICBMClientEvent,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x04_0x14_ICBMClientEvent{
				Cookie:     inBody.Cookie,
				ChannelID:  inBody.ChannelID,
				ScreenName: sess.ScreenName(),
				Event:      inBody.Event,
			},
		})
		return nil
	}
}

// EvilRequest handles user warning (a.k.a evil) notifications. It receives
// wire.ICBMEvilRequest warning SNAC, increments the warned user's warning
// level, and sends the warned user a notification informing them that they
// have been warned. The user may choose to warn anonymously or
// non-anonymously. It returns SNAC wire.ICBMEvilReply to confirm that the
// warning was sent. Users may not warn themselves or warn users they have
// blocked or are blocked by.
func (s ICBMService) EvilRequest(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x08_ICBMEvilRequest) (wire.SNACMessage, error) {
	// don't let users warn themselves, it causes the AIM client to go into a
	// weird state.
	if inBody.ScreenName == sess.ScreenName() {
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.ICBM,
				SubGroup:  wire.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: wire.ErrorCodeNotSupportedByHost,
			},
		}, nil
	}

	blocked, err := s.feedbagManager.BlockedState(sess.ScreenName(), inBody.ScreenName)
	if err != nil {
		return wire.SNACMessage{}, err
	}
	if blocked != state.BlockedNo {
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.ICBM,
				SubGroup:  wire.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: wire.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	recipSess := s.messageRelayer.RetrieveByScreenName(inBody.ScreenName)
	if recipSess == nil {
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.ICBM,
				SubGroup:  wire.ICBMErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: wire.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	increase := evilDelta
	if inBody.SendAs == 1 {
		increase = evilDeltaAnon
	}
	recipSess.IncrementWarning(increase)

	var notif any
	if inBody.SendAs == 0 {
		notif = wire.SNAC_0x01_0x10_OServiceEvilNotification{
			NewEvil: recipSess.Warning(),
			TLVUserInfo: wire.TLVUserInfo{
				ScreenName:   sess.ScreenName(),
				WarningLevel: recipSess.Warning(),
			},
		}
	} else {
		notif = wire.SNAC_0x01_0x10_OServiceEvilNotificationAnon{
			NewEvil: recipSess.Warning(),
		}
	}

	s.messageRelayer.RelayToScreenName(ctx, recipSess.ScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceEvilNotification,
		},
		Body: notif,
	})

	// inform the warned user's buddies that their warning level has increased
	if err := s.buddyUpdateBroadcaster.BroadcastBuddyArrived(ctx, recipSess); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMEvilReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x04_0x09_ICBMEvilReply{
			EvilDeltaApplied: increase,
			UpdatedEvilValue: recipSess.Warning(),
		},
	}, nil
}
