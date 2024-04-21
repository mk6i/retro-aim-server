package foodgroup

import (
	"context"
	"errors"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// omitCaps is the map of to filter out of the client's capability list
// because they are not currently supported by the server.
var omitCaps = map[[16]byte]bool{
	// 0946134a-4c7f-11d1-8222-444553540000 (games)
	{9, 70, 19, 74, 76, 127, 17, 209, 130, 34, 68, 69, 83, 84, 0, 0}: true,
	// 0946134d-4c7f-11d1-8222-444553540000 (ICQ inter-op)
	{9, 70, 19, 77, 76, 127, 17, 209, 130, 34, 68, 69, 83, 84, 0, 0}: true,
	// 09461341-4c7f-11d1-8222-444553540000 (voice chat)
	{9, 70, 19, 65, 76, 127, 17, 209, 130, 34, 68, 69, 83, 84, 0, 0}: true,
}

// NewLocateService creates a new instance of LocateService.
func NewLocateService(
	messageRelayer MessageRelayer,
	feedbagManager FeedbagManager,
	profileManager ProfileManager,
	legacyBuddyListManager LegacyBuddyListManager,
) LocateService {
	return LocateService{
		feedbagManager:         feedbagManager,
		legacyBuddyListManager: legacyBuddyListManager,
		profileManager:         profileManager,
		sessionManager:         messageRelayer,
	}
}

// LocateService provides functionality for the Locate food group, which is
// responsible for user profiles, user info lookups, directory information, and
// keyword lookups.
type LocateService struct {
	feedbagManager         FeedbagManager
	legacyBuddyListManager LegacyBuddyListManager
	profileManager         ProfileManager
	sessionManager         MessageRelayer
}

// RightsQuery returns SNAC wire.LocateRightsReply, which contains Locate food
// group settings for the current user.
func (s LocateService) RightsQuery(_ context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateRightsReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x02_0x03_LocateRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					// these are arbitrary values--AIM clients seem to perform
					// OK with them
					wire.NewTLV(wire.LocateTLVTagsRightsMaxSigLen, uint16(1000)),
					wire.NewTLV(wire.LocateTLVTagsRightsMaxCapabilitiesLen, uint16(1000)),
					wire.NewTLV(wire.LocateTLVTagsRightsMaxFindByEmailList, uint16(1000)),
					wire.NewTLV(wire.LocateTLVTagsRightsMaxCertsLen, uint16(1000)),
					wire.NewTLV(wire.LocateTLVTagsRightsMaxMaxShortCapabilities, uint16(1000)),
				},
			},
		},
	}
}

// SetInfo sets the user's profile, away message or capabilities.
func (s LocateService) SetInfo(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x02_0x04_LocateSetInfo) error {
	// update profile
	if profile, hasProfile := inBody.String(wire.LocateTLVTagsInfoSigData); hasProfile {
		if err := s.profileManager.SetProfile(sess.ScreenName(), profile); err != nil {
			return err
		}
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := inBody.String(wire.LocateTLVTagsInfoUnavailableData); hasAwayMsg {
		sess.SetAwayMessage(awayMsg)
		if err := broadcastArrival(ctx, sess, s.sessionManager, s.feedbagManager, s.legacyBuddyListManager); err != nil {
			return err
		}
	}

	// update client capabilities (buddy icon, chat, etc...)
	if b, hasCaps := inBody.Slice(wire.LocateTLVTagsInfoCapabilities); hasCaps {
		if len(b)%16 != 0 {
			return errors.New("capability list must be array of 16-byte values")
		}
		var caps [][16]byte
		for i := 0; i < len(b); i += 16 {
			var c [16]byte
			copy(c[:], b[i:i+16])
			if _, found := omitCaps[c]; found {
				continue
			}
			caps = append(caps, c)
		}
		sess.SetCaps(caps)
	}

	return nil
}

// UserInfoQuery fetches display information about an arbitrary user (not the
// current user). It returns wire.LocateUserInfoReply, which contains the
// profile, if requested, and/or the away message, if requested. This is a v2
// of UserInfoQuery.
func (s LocateService) UserInfoQuery(_ context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x05_LocateUserInfoQuery) (wire.SNACMessage, error) {
	blocked, err := s.feedbagManager.BlockedState(sess.ScreenName(), inBody.ScreenName)
	switch {
	case err != nil:
		return wire.SNACMessage{}, err
	case blocked != state.BlockedNo:
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.Locate,
				SubGroup:  wire.LocateErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: wire.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	buddySess := s.sessionManager.RetrieveByScreenName(inBody.ScreenName)
	if buddySess == nil {
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.Locate,
				SubGroup:  wire.LocateErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: wire.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	var list wire.TLVList

	if inBody.RequestProfile() {
		profile, err := s.profileManager.Profile(inBody.ScreenName)
		if err != nil {
			return wire.SNACMessage{}, err
		}
		list.AppendList([]wire.TLV{
			wire.NewTLV(wire.LocateTLVTagsInfoSigMime, `text/aolrtf; charset="us-ascii"`),
			wire.NewTLV(wire.LocateTLVTagsInfoSigData, profile),
		})
	}

	if inBody.RequestAwayMessage() {
		list.AppendList([]wire.TLV{
			wire.NewTLV(wire.LocateTLVTagsInfoUnavailableMime, `text/aolrtf; charset="us-ascii"`),
			wire.NewTLV(wire.LocateTLVTagsInfoUnavailableData, buddySess.AwayMessage()),
		})
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateUserInfoReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x02_0x06_LocateUserInfoReply{
			TLVUserInfo: buddySess.TLVUserInfo(),
			LocateInfo: wire.TLVRestBlock{
				TLVList: list,
			},
		},
	}, nil
}

// SetDirInfo sets directory information for current user (first name, last
// name, etc). This method does nothing and exists to placate the AIM client.
// It returns wire.LocateSetDirReply with a canned success message.
func (s LocateService) SetDirInfo(_ context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateSetDirReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x02_0x0A_LocateSetDirReply{
			Result: 1,
		},
	}
}

// SetKeywordInfo sets profile keywords and interests. This method does nothing
// and exists to placate the AIM client. It returns wire.LocateSetKeywordReply
// with a canned success message.
func (s LocateService) SetKeywordInfo(_ context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateSetKeywordReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x02_0x10_LocateSetKeywordReply{
			Unknown: 1,
		},
	}
}
