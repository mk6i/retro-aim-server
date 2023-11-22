package handler

import (
	"context"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

func NewLocateService(sm SessionManager, fm FeedbagManager, pm ProfileManager) LocateService {
	return LocateService{
		sessionManager: sm,
		feedbagManager: fm,
		profileManager: pm,
	}
}

type LocateService struct {
	sessionManager SessionManager
	feedbagManager FeedbagManager
	profileManager ProfileManager
}

func (s LocateService) RightsQueryHandler(context.Context) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateRightsReply,
		},
		Body: oscar.SNAC_0x02_0x03_LocateRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(0x01, uint16(1000)),
					oscar.NewTLV(0x02, uint16(1000)),
					oscar.NewTLV(0x03, uint16(1000)),
					oscar.NewTLV(0x04, uint16(1000)),
					oscar.NewTLV(0x05, uint16(1000)),
				},
			},
		},
	}
}

func (s LocateService) SetInfoHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error {
	// update profile
	if profile, hasProfile := snacPayloadIn.GetString(oscar.LocateTLVTagsInfoSigData); hasProfile {
		if err := s.profileManager.UpsertProfile(sess.ScreenName(), profile); err != nil {
			return err
		}
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := snacPayloadIn.GetString(oscar.LocateTLVTagsInfoUnavailableData); hasAwayMsg {
		sess.SetAwayMessage(awayMsg)
		if err := broadcastArrival(ctx, sess, s.sessionManager, s.feedbagManager); err != nil {
			return err
		}
	}
	return nil
}

func (s LocateService) UserInfoQuery2Handler(_ context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (oscar.SNACMessage, error) {
	blocked, err := s.feedbagManager.Blocked(sess.ScreenName(), snacPayloadIn.ScreenName)
	switch {
	case err != nil:
		return oscar.SNACMessage{}, err
	case blocked != state.BlockedNo:
		return oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.Locate,
				SubGroup:  oscar.LocateErr,
			},
			Body: oscar.SNACError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	buddySess := s.sessionManager.RetrieveByScreenName(snacPayloadIn.ScreenName)
	if buddySess == nil {
		return oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.Locate,
				SubGroup:  oscar.LocateErr,
			},
			Body: oscar.SNACError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	var list oscar.TLVList

	if snacPayloadIn.RequestProfile() {
		profile, err := s.profileManager.RetrieveProfile(snacPayloadIn.ScreenName)
		if err != nil {
			return oscar.SNACMessage{}, err
		}
		list.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.LocateTLVTagsInfoSigMime, `text/aolrtf; charset="us-ascii"`),
			oscar.NewTLV(oscar.LocateTLVTagsInfoSigData, profile),
		})
	}

	if snacPayloadIn.RequestAwayMessage() {
		list.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableMime, `text/aolrtf; charset="us-ascii"`),
			oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableData, buddySess.AwayMessage()),
		})
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateUserInfoReply,
		},
		Body: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
			TLVUserInfo: buddySess.TLVUserInfo(),
			LocateInfo: oscar.TLVRestBlock{
				TLVList: list,
			},
		},
	}, nil
}

func (s LocateService) SetDirInfoHandler(_ context.Context) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateSetDirReply,
		},
		Body: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
			Result: 1,
		},
	}
}

func (s LocateService) SetKeywordInfoHandler(_ context.Context) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateSetKeywordReply,
		},
		Body: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
			Unknown: 1,
		},
	}
}
