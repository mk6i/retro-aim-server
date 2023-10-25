package server

import (
	"errors"
	"io"

	"github.com/mkaminski/goaim/oscar"
)

type LocateHandler interface {
	RightsQueryHandler() XMessage
	SetDirInfoHandler() XMessage
	SetInfoHandler(sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error
	SetKeywordInfoHandler() XMessage
	UserInfoQuery2Handler(sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (XMessage, error)
}

func NewLocateRouter() LocateRouter {
	return LocateRouter{
		LocateHandler: LocateService{},
	}
}

type LocateRouter struct {
	LocateHandler
}

func (rt LocateRouter) RouteLocate(sess *Session, sm SessionManager, fm *FeedbagStore, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.SubGroup {
	case oscar.LocateRightsQuery:
		outSNAC := rt.RightsQueryHandler()
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.LocateSetInfo:
		snacPayloadIn := oscar.SNAC_0x02_0x04_LocateSetInfo{}
		if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
			return err
		}
		return rt.SetInfoHandler(sess, sm, fm, fm, snacPayloadIn)
	case oscar.LocateSetDirInfo:
		snacPayloadIn := oscar.SNAC_0x02_0x09_LocateSetDirInfo{}
		if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
			return err
		}
		outSNAC := rt.SetDirInfoHandler()
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.LocateGetDirInfo:
		snacPayloadIn := oscar.SNAC_0x02_0x0B_LocateGetDirInfo{}
		return oscar.Unmarshal(&snacPayloadIn, r)
	case oscar.LocateSetKeywordInfo:
		snacPayloadIn := oscar.SNAC_0x02_0x0F_LocateSetKeywordInfo{}
		if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
			return err
		}
		outSNAC := rt.SetKeywordInfoHandler()
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.LocateUserInfoQuery2:
		snacPayloadIn := oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{}
		if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
			return err
		}
		outSNAC, err := rt.UserInfoQuery2Handler(sess, sm, fm, fm, snacPayloadIn)
		if err != nil {
			return err
		}
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type LocateService struct {
}

func (s LocateService) RightsQueryHandler() XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: LOCATE,
			SubGroup:  oscar.LocateRightsReply,
		},
		snacOut: oscar.SNAC_0x02_0x03_LocateRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					{
						TType: 0x01,
						Val:   uint16(1000),
					},
					{
						TType: 0x02,
						Val:   uint16(1000),
					},
					{
						TType: 0x03,
						Val:   uint16(1000),
					},
					{
						TType: 0x04,
						Val:   uint16(1000),
					},
					{
						TType: 0x05,
						Val:   uint16(1000),
					},
				},
			},
		},
	}
}

func (s LocateService) SetInfoHandler(sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error {
	// update profile
	if profile, hasProfile := snacPayloadIn.GetString(oscar.LocateTLVTagsInfoSigData); hasProfile {
		if err := pm.UpsertProfile(sess.ScreenName, profile); err != nil {
			return err
		}
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := snacPayloadIn.GetString(oscar.LocateTLVTagsInfoUnavailableData); hasAwayMsg {
		sess.SetAwayMessage(awayMsg)
		if err := NotifyArrival(sess, sm, fm); err != nil {
			return err
		}
	}
	return nil
}

func (s LocateService) UserInfoQuery2Handler(sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (XMessage, error) {
	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)
	switch {
	case err != nil:
		return XMessage{}, nil
	case blocked != BlockedNo:
		return XMessage{
			snacFrame: oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  oscar.LocateErr,
			},
			snacOut: oscar.SnacError{
				Code: ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	buddySess, err := sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	switch {
	case errors.Is(err, ErrSessNotFound):
		return XMessage{
			snacFrame: oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  oscar.LocateErr,
			},
			snacOut: oscar.SnacError{
				Code: ErrorCodeNotLoggedOn,
			},
		}, nil
	case err != nil:
		return XMessage{}, err
	}

	var list oscar.TLVList

	if snacPayloadIn.RequestProfile() {
		profile, err := pm.RetrieveProfile(snacPayloadIn.ScreenName)
		if err != nil {
			return XMessage{}, err
		}
		list.AddTLVList([]oscar.TLV{
			{
				TType: oscar.LocateTLVTagsInfoSigMime,
				Val:   `text/aolrtf; charset="us-ascii"`,
			},
			{
				TType: oscar.LocateTLVTagsInfoSigData,
				Val:   profile,
			},
		})
	}

	if snacPayloadIn.RequestAwayMessage() {
		list.AddTLVList([]oscar.TLV{
			{
				TType: oscar.LocateTLVTagsInfoUnavailableMime,
				Val:   `text/aolrtf; charset="us-ascii"`,
			},
			{
				TType: oscar.LocateTLVTagsInfoUnavailableData,
				Val:   buddySess.GetAwayMessage(),
			},
		})
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: LOCATE,
			SubGroup:  oscar.LocateUserInfoReply,
		},
		snacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
			TLVUserInfo: buddySess.GetTLVUserInfo(),
			LocateInfo: oscar.TLVRestBlock{
				TLVList: list,
			},
		},
	}, nil
}

func (s LocateService) SetDirInfoHandler() XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: LOCATE,
			SubGroup:  oscar.LocateSetDirReply,
		},
		snacOut: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
			Result: 1,
		},
	}
}

func (s LocateService) SetKeywordInfoHandler() XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: LOCATE,
			SubGroup:  oscar.LocateSetKeywordReply,
		},
		snacOut: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
			Unknown: 1,
		},
	}
}
