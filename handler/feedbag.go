package handler

import (
	"context"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

// NewFeedbagService creates a new instance of FeedbagService.
func NewFeedbagService(messageRelayer MessageRelayer, feedbagManager FeedbagManager) FeedbagService {
	return FeedbagService{
		messageRelayer: messageRelayer,
		feedbagManager: feedbagManager,
	}
}

// FeedbagService provides handlers for the Feedbag food group.
type FeedbagService struct {
	messageRelayer MessageRelayer
	feedbagManager FeedbagManager
}

// RightsQueryHandler returns SNAC oscar.FeedbagRightsReply, which contains
// Feedbag food group settings for the current user. The values within the SNAC
// are not well understood but seem to make the AIM client happy.
func (s FeedbagService) RightsQueryHandler(_ context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagRightsReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x13_0x03_FeedbagRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.FeedbagRightsMaxItemAttrs, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxItemsByClass, []uint16{
						0x3D, // max num of contacts
						0x3D, // max num of groups
						0x64, // max visible contacts
						0x64, // max invisible contacts
						0x01, // max vis/invis bitmasks
						0x01, // max presense info fields
						0x32, // limit for item type 06
						0x00, // limit for item type 07
						0x00, // limit for item type 08
						0x03, // limit for item type 09
						0x00, // limit for item type 0a
						0x00, // limit for item type 0b
						0x00, // limit for item type 0c
						0x80, // limit for item type 0d
						0xFF, // max ignore list entries
						0x14, // limit for item type 0f
						0xC8, // limit for item 10
						0x01, // limit for item 11
						0x00, // limit for item 12
						0x01, // limit for item 13
						0x00, // limit for item 14
					}),
					oscar.NewTLV(oscar.FeedbagRightsMaxClientItems, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxItemNameLen, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxRecentBuddies, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsInteractionBuddies, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsInteractionHalfLife, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsInteractionMaxScore, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxBuddiesPerGroup, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxMegaBots, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxSmartGroups, uint16(100)),
				},
			},
		},
	}
}

// QueryHandler fetches the user's feedbag (aka buddy list). It returns
// oscar.FeedbagReply, which contains feedbag entries.
func (s FeedbagService) QueryHandler(_ context.Context, sess *state.Session, inFrame oscar.SNACFrame) (oscar.SNACMessage, error) {
	fb, err := s.feedbagManager.Retrieve(sess.ScreenName())
	if err != nil {
		return oscar.SNACMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = s.feedbagManager.LastModified(sess.ScreenName())
		if err != nil {
			return oscar.SNACMessage{}, err
		}
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

// QueryIfModifiedHandler fetches the user's feedbag (aka buddy list). It
// returns oscar.FeedbagReplyNotModified if the feedbag was last modified
// before inBody.LastUpdate, else return oscar.FeedbagReply, which contains
// feedbag entries.
func (s FeedbagService) QueryIfModifiedHandler(_ context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x05_FeedbagQueryIfModified) (oscar.SNACMessage, error) {
	fb, err := s.feedbagManager.Retrieve(sess.ScreenName())
	if err != nil {
		return oscar.SNACMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = s.feedbagManager.LastModified(sess.ScreenName())
		if err != nil {
			return oscar.SNACMessage{}, err
		}
		if lm.Before(time.Unix(int64(inBody.LastUpdate), 0)) {
			return oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReplyNotModified,
					RequestID: inFrame.RequestID,
				},
				Body: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(lm.Unix()),
					Count:      uint8(len(fb)),
				},
			}, nil
		}
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

// InsertItemHandler adds items to the user's feedbag (aka buddy list). Sends
// user buddy arrival notifications for each online & visible buddy added to
// the feedbag. Sends a buddy departure notification to blocked buddies if
// current user is visible. It returns oscar.FeedbagStatus, which contains
// insert confirmation.
func (s FeedbagService) InsertItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x08_FeedbagInsertItem) (oscar.SNACMessage, error) {
	for _, item := range inBody.Items {
		// don't let users block themselves, it causes the AIM client to go
		// into a weird state.
		if item.ClassID == 3 && item.Name == sess.ScreenName() {
			return oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagErr,
					RequestID: inFrame.RequestID,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeNotSupportedByHost,
				},
			}, nil
		}
	}

	if err := s.feedbagManager.Upsert(sess.ScreenName(), inBody.Items); err != nil {
		return oscar.SNACMessage{}, nil
	}

	for _, item := range inBody.Items {
		switch item.ClassID {
		case oscar.FeedbagClassIdBuddy, oscar.FeedbagClassIDPermit: // add new buddy
			buddy := s.messageRelayer.RetrieveByScreenName(item.Name)
			if buddy == nil || buddy.Invisible() {
				continue
			}
			unicastArrival(ctx, buddy, sess, s.messageRelayer)
		case oscar.FeedbagClassIDDeny: // block buddy
			if sess.Invisible() {
				continue // user's offline, don't send departure notification
			}
			blockedSess := s.messageRelayer.RetrieveByScreenName(item.Name)
			if blockedSess == nil {
				continue // blocked buddy is offline, nothing to do here
			}
			// alert blocked buddy that current user is offline
			unicastDeparture(ctx, sess, blockedSess, s.messageRelayer)
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range inBody.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagStatus,
			RequestID: inFrame.RequestID,
		},
		Body: snacPayloadOut,
	}, nil
}

// UpdateItemHandler updates items in the user's feedbag (aka buddy list).
// Sends user buddy arrival notifications for each online & visible buddy added
// to the feedbag. It returns oscar.FeedbagStatus, which contains update
// confirmation.
func (s FeedbagService) UpdateItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x09_FeedbagUpdateItem) (oscar.SNACMessage, error) {
	if err := s.feedbagManager.Upsert(sess.ScreenName(), inBody.Items); err != nil {
		return oscar.SNACMessage{}, nil
	}

	for _, item := range inBody.Items {
		switch item.ClassID {
		case oscar.FeedbagClassIdBuddy, oscar.FeedbagClassIDPermit:
			buddy := s.messageRelayer.RetrieveByScreenName(item.Name)
			if buddy == nil || buddy.Invisible() {
				continue
			}
			unicastArrival(ctx, buddy, sess, s.messageRelayer)
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range inBody.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagStatus,
			RequestID: inFrame.RequestID,
		},
		Body: snacPayloadOut,
	}, nil
}

// DeleteItemHandler removes items from feedbag (aka buddy list). Sends user
// buddy arrival notifications for each online & visible buddy added to
// the feedbag. Sends buddy arrival notifications to each unblocked buddy if
// current user is visible. It returns oscar.FeedbagStatus, which contains update
// confirmation.
func (s FeedbagService) DeleteItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x0A_FeedbagDeleteItem) (oscar.SNACMessage, error) {
	if err := s.feedbagManager.Delete(sess.ScreenName(), inBody.Items); err != nil {
		return oscar.SNACMessage{}, err
	}

	for _, item := range inBody.Items {
		if item.ClassID == oscar.FeedbagClassIDDeny {
			unblockedSess := s.messageRelayer.RetrieveByScreenName(item.Name)
			if unblockedSess == nil {
				continue // unblocked user is offline, nothing to do here
			}
			if !sess.Invisible() {
				// alert unblocked user that current user is online
				unicastArrival(ctx, sess, unblockedSess, s.messageRelayer)
			}
			if !unblockedSess.Invisible() {
				// alert current user that unblocked user is online
				unicastArrival(ctx, unblockedSess, sess, s.messageRelayer)
			}
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range inBody.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000) // success by default
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagStatus,
			RequestID: inFrame.RequestID,
		},
		Body: snacPayloadOut,
	}, nil
}

// StartClusterHandler exists to capture the SNAC input in unit tests to verify
// it's correctly unmarshalled.
func (s FeedbagService) StartClusterHandler(context.Context, oscar.SNACFrame, oscar.SNAC_0x13_0x11_FeedbagStartCluster) {
}
