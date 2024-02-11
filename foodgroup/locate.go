package foodgroup

import (
	"context"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewLocateService creates a new instance of LocateService.
func NewLocateService(messageRelayer MessageRelayer, feedbagManager FeedbagManager, profileManager ProfileManager) LocateService {
	return LocateService{
		sessionManager: messageRelayer,
		feedbagManager: feedbagManager,
		profileManager: profileManager,
	}
}

// LocateService provides functionality for the Locate food group, which is
// responsible for user profiles, user info lookups, directory information, and
// keyword lookups.
type LocateService struct {
	sessionManager MessageRelayer
	feedbagManager FeedbagManager
	profileManager ProfileManager
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

// SetInfo sets the user's profile or away message.
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
		if err := broadcastArrival(ctx, sess, s.sessionManager, s.feedbagManager); err != nil {
			return err
		}
	}
	return nil
}

// UserInfoQuery2 fetches display information about an arbitrary user (not the
// current user). It returns wire.LocateUserInfoReply, which contains the
// profile, if requested, and/or the away message, if requested. This is a v2
// of UserInfoQuery.
func (s LocateService) UserInfoQuery2(_ context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x15_LocateUserInfoQuery2) (wire.SNACMessage, error) {
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
