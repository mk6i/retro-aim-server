package foodgroup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewFeedbagService creates a new instance of FeedbagService.
func NewFeedbagService(logger *slog.Logger, messageRelayer MessageRelayer, feedbagManager FeedbagManager, bartManager BARTManager) FeedbagService {
	return FeedbagService{
		bartManager:    bartManager,
		feedbagManager: feedbagManager,
		logger:         logger,
		messageRelayer: messageRelayer,
	}
}

// FeedbagService provides functionality for the Feedbag food group, which
// handles buddy list management.
type FeedbagService struct {
	messageRelayer MessageRelayer
	feedbagManager FeedbagManager
	bartManager    BARTManager
	logger         *slog.Logger
}

// RightsQuery returns SNAC wire.FeedbagRightsReply, which contains Feedbag
// food group settings for the current user. The values within the SNAC are not
// well understood but seem to make the AIM client happy.
func (s FeedbagService) RightsQuery(_ context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	// maxItemsByClass defines per-type item limits. Types not listed here are
	// 0 by default. The slice size is equal to the maximum "enum" value+1.
	maxItemsByClass := make([]uint16, 21)
	maxItemsByClass[wire.FeedbagClassIdBuddy] = 61
	maxItemsByClass[wire.FeedbagClassIdGroup] = 61
	maxItemsByClass[wire.FeedbagClassIDPermit] = 100
	maxItemsByClass[wire.FeedbagClassIDDeny] = 100
	maxItemsByClass[wire.FeedbagClassIdPdinfo] = 1
	maxItemsByClass[wire.FeedbagClassIdBuddyPrefs] = 1
	maxItemsByClass[wire.FeedbagClassIdNonbuddy] = 50
	maxItemsByClass[wire.FeedbagClassIdClientPrefs] = 3
	maxItemsByClass[wire.FeedbagClassIdWatchList] = 128
	maxItemsByClass[wire.FeedbagClassIdIgnoreList] = 255
	maxItemsByClass[wire.FeedbagClassIdDateTime] = 20
	maxItemsByClass[wire.FeedbagClassIdExternalUser] = 200
	maxItemsByClass[wire.FeedbagClassIdRootCreator] = 1
	maxItemsByClass[wire.FeedbagClassIdImportTimestamp] = 1
	maxItemsByClass[wire.FeedbagClassIdBart] = 200

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagRightsReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x13_0x03_FeedbagRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.FeedbagRightsMaxItemAttrs, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxItemsByClass, maxItemsByClass),
					wire.NewTLV(wire.FeedbagRightsMaxClientItems, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxItemNameLen, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxRecentBuddies, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsInteractionBuddies, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsInteractionHalfLife, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsInteractionMaxScore, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxBuddiesPerGroup, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxMegaBots, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxSmartGroups, uint16(100)),
				},
			},
		},
	}
}

// Query fetches the user's feedbag (aka buddy list). It returns
// wire.FeedbagReply, which contains feedbag entries.
func (s FeedbagService) Query(_ context.Context, sess *state.Session, inFrame wire.SNACFrame) (wire.SNACMessage, error) {
	fb, err := s.feedbagManager.Feedbag(sess.ScreenName())
	if err != nil {
		return wire.SNACMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = s.feedbagManager.FeedbagLastModified(sess.ScreenName())
		if err != nil {
			return wire.SNACMessage{}, err
		}
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

// QueryIfModified fetches the user's feedbag (aka buddy list). It returns
// wire.FeedbagReplyNotModified if the feedbag was last modified before
// inBody.LastUpdate, else return wire.FeedbagReply, which contains feedbag
// entries.
func (s FeedbagService) QueryIfModified(_ context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x13_0x05_FeedbagQueryIfModified) (wire.SNACMessage, error) {
	fb, err := s.feedbagManager.Feedbag(sess.ScreenName())
	if err != nil {
		return wire.SNACMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = s.feedbagManager.FeedbagLastModified(sess.ScreenName())
		if err != nil {
			return wire.SNACMessage{}, err
		}
		if lm.Before(time.Unix(int64(inBody.LastUpdate), 0)) {
			return wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagReplyNotModified,
					RequestID: inFrame.RequestID,
				},
				Body: wire.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(lm.Unix()),
					Count:      uint8(len(fb)),
				},
			}, nil
		}
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

// UpsertItem yadada
// InsertItem adds items to the user's feedbag (aka buddy list). Sends user
// buddy arrival notifications for each online & visible buddy added to the
// feedbag. Sends a buddy departure notification to blocked buddies if current
// user is visible. It returns wire.FeedbagStatus, which contains insert
// confirmation.
// UpdateItem updates items in the user's feedbag (aka buddy list). Sends user
// buddy arrival notifications for each online & visible buddy added to the
// feedbag. It returns wire.FeedbagStatus, which contains update confirmation.
func (s FeedbagService) UpsertItem(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, items []wire.FeedbagItem) (wire.SNACMessage, error) {
	for _, item := range items {
		// don't let users block themselves, it causes the AIM client to go
		// into a weird state.
		if item.ClassID == 3 && item.Name == sess.ScreenName() {
			return wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagErr,
					RequestID: inFrame.RequestID,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotSupportedByHost,
				},
			}, nil
		}
	}

	if err := s.feedbagManager.FeedbagUpsert(sess.ScreenName(), items); err != nil {
		return wire.SNACMessage{}, nil
	}

	for _, item := range items {
		switch item.ClassID {
		case wire.FeedbagClassIdBuddy, wire.FeedbagClassIDPermit: // add new buddy
			buddy := s.messageRelayer.RetrieveByScreenName(item.Name)
			if buddy == nil || buddy.Invisible() {
				continue
			}
			if err := unicastArrival(ctx, buddy, sess, s.messageRelayer, s.feedbagManager); err != nil {
				return wire.SNACMessage{}, nil
			}
		case wire.FeedbagClassIDDeny: // block buddy
			if sess.Invisible() {
				continue // user's offline, don't send departure notification
			}
			blockedSess := s.messageRelayer.RetrieveByScreenName(item.Name)
			if blockedSess == nil {
				continue // blocked buddy is offline, nothing to do here
			}
			// alert blocked buddy that current user is offline
			unicastDeparture(ctx, sess, blockedSess, s.messageRelayer)
		case wire.FeedbagClassIdBart:
			if err := s.broadcastIconUpdate(ctx, sess, item); err != nil {
				return wire.SNACMessage{}, err
			}
		}
	}

	snacPayloadOut := wire.SNAC_0x13_0x0E_FeedbagStatus{}
	for range items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagStatus,
			RequestID: inFrame.RequestID,
		},
		Body: snacPayloadOut,
	}, nil
}

// broadcastIconUpdate informs clients about buddy icon update. If the BART
// store doesn't have the icon, then tell the client to upload the buddy icon.
// If the icon already exists, tell the user's buddies about the icon change.
func (s FeedbagService) broadcastIconUpdate(ctx context.Context, sess *state.Session, item wire.FeedbagItem) error {
	btlv := wire.BARTInfo{}
	if b, hasBuf := item.Slice(wire.FeedbagAttributesBartInfo); hasBuf {
		if err := wire.Unmarshal(&btlv, bytes.NewBuffer(b)); err != nil {
			return err
		}
	} else {
		return errors.New("unable to extract icon payload")
	}

	if bytes.Equal(btlv.Hash, wire.GetClearIconHash()) {
		s.logger.DebugContext(ctx, "user is clearing icon",
			"hash", fmt.Sprintf("%x", btlv.Hash))
		// tell buddies about the icon update
		return broadcastArrival(ctx, sess, s.messageRelayer, s.feedbagManager)
	}

	bid := wire.BARTID{
		Type: wire.BARTTypesBuddyIcon,
		BARTInfo: wire.BARTInfo{
			Flags: wire.BARTFlagsCustom,
			Hash:  btlv.Hash,
		},
	}
	if b, err := s.bartManager.BARTRetrieve(btlv.Hash); err != nil {
		return err
	} else if len(b) == 0 {
		// icon doesn't exist, tell the client to upload buddy icon
		s.logger.DebugContext(ctx, "icon doesn't exist in BART store, client must upload the icon file",
			"hash", fmt.Sprintf("%x", btlv.Hash))
		bid.Flags |= wire.BARTFlagsUnknown
	} else {
		s.logger.DebugContext(ctx, "icon already exists in BART store, don't upload the icon file",
			"hash", fmt.Sprintf("%x", btlv.Hash))
		// tell buddies about the icon update
		if err := broadcastArrival(ctx, sess, s.messageRelayer, s.feedbagManager); err != nil {
			return err
		}
	}

	s.messageRelayer.RelayToScreenName(ctx, sess.ScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceBartReply,
		},
		Body: wire.SNAC_0x01_0x21_OServiceBARTReply{
			BARTID: bid,
		},
	})

	return nil
}

// DeleteItem removes items from feedbag (aka buddy list). Sends user buddy
// arrival notifications for each online & visible buddy added to the feedbag.
// Sends buddy arrival notifications to each unblocked buddy if current user is
// visible. It returns wire.FeedbagStatus, which contains update confirmation.
func (s FeedbagService) DeleteItem(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x13_0x0A_FeedbagDeleteItem) (wire.SNACMessage, error) {
	if err := s.feedbagManager.FeedbagDelete(sess.ScreenName(), inBody.Items); err != nil {
		return wire.SNACMessage{}, err
	}

	for _, item := range inBody.Items {
		if item.ClassID == wire.FeedbagClassIDDeny {
			unblockedSess := s.messageRelayer.RetrieveByScreenName(item.Name)
			if unblockedSess == nil {
				continue // unblocked user is offline, nothing to do here
			}
			if !sess.Invisible() {
				// alert unblocked user that current user is online
				if err := unicastArrival(ctx, sess, unblockedSess, s.messageRelayer, s.feedbagManager); err != nil {
					return wire.SNACMessage{}, err
				}
			}
			if !unblockedSess.Invisible() {
				// alert current user that unblocked user is online
				if err := unicastArrival(ctx, unblockedSess, sess, s.messageRelayer, s.feedbagManager); err != nil {
					return wire.SNACMessage{}, err
				}
			}
		}
	}

	snacPayloadOut := wire.SNAC_0x13_0x0E_FeedbagStatus{}
	for range inBody.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000) // success by default
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagStatus,
			RequestID: inFrame.RequestID,
		},
		Body: snacPayloadOut,
	}, nil
}

// StartCluster exists to capture the SNAC input in unit tests to verify it's
// correctly unmarshalled.
func (s FeedbagService) StartCluster(context.Context, wire.SNACFrame, wire.SNAC_0x13_0x11_FeedbagStartCluster) {
}
