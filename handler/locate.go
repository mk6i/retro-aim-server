package handler

import (
	"context"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

// NewLocateService creates a new instance of LocateService.
func NewLocateService(messageRelayer MessageRelayer, feedbagManager FeedbagManager, profileManager ProfileManager) LocateService {
	return LocateService{
		sessionManager: messageRelayer,
		feedbagManager: feedbagManager,
		profileManager: profileManager,
	}
}

// LocateService provides handlers for the Locate food group.
type LocateService struct {
	sessionManager MessageRelayer
	feedbagManager FeedbagManager
	profileManager ProfileManager
}

// RightsQueryHandler returns SNAC oscar.LocateRightsReply, which contains
// Locate food group settings for the current user.
func (s LocateService) RightsQueryHandler(_ context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateRightsReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x02_0x03_LocateRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					// these are arbitrary values--AIM clients seem to perform
					// OK with them
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxSigLen, uint16(1000)),
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxCapabilitiesLen, uint16(1000)),
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxFindByEmailList, uint16(1000)),
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxCertsLen, uint16(1000)),
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxMaxShortCapabilities, uint16(1000)),
				},
			},
		},
	}
}

// SetInfoHandler sets the user's profile or away message.
func (s LocateService) SetInfoHandler(ctx context.Context, sess *state.Session, inBody oscar.SNAC_0x02_0x04_LocateSetInfo) error {
	// update profile
	if profile, hasProfile := inBody.GetString(oscar.LocateTLVTagsInfoSigData); hasProfile {
		if err := s.profileManager.UpsertProfile(sess.ScreenName(), profile); err != nil {
			return err
		}
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := inBody.GetString(oscar.LocateTLVTagsInfoUnavailableData); hasAwayMsg {
		sess.SetAwayMessage(awayMsg)
		if err := broadcastArrival(ctx, sess, s.sessionManager, s.feedbagManager); err != nil {
			return err
		}
	}
	return nil
}

// UserInfoQuery2Handler fetches display information about an arbitrary user
// (not the current user). It returns oscar.LocateUserInfoReply, which contains
// the profile, if requested, and/or the away message, if requested. This is a
// v2 of the UserInfoQueryHandler handler.
func (s LocateService) UserInfoQuery2Handler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (oscar.SNACMessage, error) {
	blocked, err := s.feedbagManager.Blocked(sess.ScreenName(), inBody.ScreenName)
	switch {
	case err != nil:
		return oscar.SNACMessage{}, err
	case blocked != state.BlockedNo:
		return oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.Locate,
				SubGroup:  oscar.LocateErr,
				RequestID: inFrame.RequestID,
			},
			Body: oscar.SNACError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	buddySess := s.sessionManager.RetrieveByScreenName(inBody.ScreenName)
	if buddySess == nil {
		return oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.Locate,
				SubGroup:  oscar.LocateErr,
				RequestID: inFrame.RequestID,
			},
			Body: oscar.SNACError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	var list oscar.TLVList

	if inBody.RequestProfile() {
		profile, err := s.profileManager.RetrieveProfile(inBody.ScreenName)
		if err != nil {
			return oscar.SNACMessage{}, err
		}
		list.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.LocateTLVTagsInfoSigMime, `text/aolrtf; charset="us-ascii"`),
			oscar.NewTLV(oscar.LocateTLVTagsInfoSigData, profile),
		})
	}

	if inBody.RequestAwayMessage() {
		list.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableMime, `text/aolrtf; charset="us-ascii"`),
			oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableData, buddySess.AwayMessage()),
		})
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateUserInfoReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
			TLVUserInfo: buddySess.TLVUserInfo(),
			LocateInfo: oscar.TLVRestBlock{
				TLVList: list,
			},
		},
	}, nil
}

// SetDirInfoHandler sets directory information for current user (first name,
// last name, etc). This method does nothing and exists to placate the AIM
// client. It returns oscar.LocateSetDirReply with a canned success message.
func (s LocateService) SetDirInfoHandler(_ context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateSetDirReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
			Result: 1,
		},
	}
}

// SetKeywordInfoHandler sets profile keywords and interests. This method does
// nothing and exists to placate the AIM client. It returns
// oscar.LocateSetKeywordReply with a canned success message.
func (s LocateService) SetKeywordInfoHandler(_ context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateSetKeywordReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
			Unknown: 1,
		},
	}
}
