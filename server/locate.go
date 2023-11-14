package server

import (
	"context"
	"github.com/mkaminski/goaim/user"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
)

type LocateHandler interface {
	RightsQueryHandler(ctx context.Context) oscar.XMessage
	SetDirInfoHandler(ctx context.Context) oscar.XMessage
	SetInfoHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error
	SetKeywordInfoHandler(ctx context.Context) oscar.XMessage
	UserInfoQuery2Handler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (oscar.XMessage, error)
}

func NewLocateRouter(logger *slog.Logger, sm SessionManager, fm FeedbagManager, pm ProfileManager) LocateRouter {
	return LocateRouter{
		LocateHandler: LocateService{
			sm: sm,
			fm: fm,
			pm: pm,
		},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type LocateRouter struct {
	LocateHandler
	RouteLogger
}

func (rt LocateRouter) RouteLocate(ctx context.Context, sess *user.Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.LocateRightsQuery:
		outSNAC := rt.RightsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.LocateSetInfo:
		inSNAC := oscar.SNAC_0x02_0x04_LocateSetInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.SetInfoHandler(ctx, sess, inSNAC)
	case oscar.LocateSetDirInfo:
		inSNAC := oscar.SNAC_0x02_0x09_LocateSetDirInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.SetDirInfoHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.LocateGetDirInfo:
		inSNAC := oscar.SNAC_0x02_0x0B_LocateGetDirInfo{}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return oscar.Unmarshal(&inSNAC, r)
	case oscar.LocateSetKeywordInfo:
		inSNAC := oscar.SNAC_0x02_0x0F_LocateSetKeywordInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.SetKeywordInfoHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.LocateUserInfoQuery2:
		inSNAC := oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.UserInfoQuery2Handler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type LocateService struct {
	sm SessionManager
	fm FeedbagManager
	pm ProfileManager
}

func (s LocateService) RightsQueryHandler(context.Context) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.LOCATE,
			SubGroup:  oscar.LocateRightsReply,
		},
		SnacOut: oscar.SNAC_0x02_0x03_LocateRightsReply{
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

func (s LocateService) SetInfoHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error {
	// update profile
	if profile, hasProfile := snacPayloadIn.GetString(oscar.LocateTLVTagsInfoSigData); hasProfile {
		if err := s.pm.UpsertProfile(sess.ScreenName(), profile); err != nil {
			return err
		}
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := snacPayloadIn.GetString(oscar.LocateTLVTagsInfoUnavailableData); hasAwayMsg {
		sess.SetAwayMessage(awayMsg)
		if err := BroadcastArrival(ctx, sess, s.sm, s.fm); err != nil {
			return err
		}
	}
	return nil
}

func (s LocateService) UserInfoQuery2Handler(_ context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (oscar.XMessage, error) {
	blocked, err := s.fm.Blocked(sess.ScreenName(), snacPayloadIn.ScreenName)
	switch {
	case err != nil:
		return oscar.XMessage{}, err
	case blocked != user.BlockedNo:
		return oscar.XMessage{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.LOCATE,
				SubGroup:  oscar.LocateErr,
			},
			SnacOut: oscar.SnacError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	buddySess := s.sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	if buddySess == nil {
		return oscar.XMessage{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.LOCATE,
				SubGroup:  oscar.LocateErr,
			},
			SnacOut: oscar.SnacError{
				Code: oscar.ErrorCodeNotLoggedOn,
			},
		}, nil
	}

	var list oscar.TLVList

	if snacPayloadIn.RequestProfile() {
		profile, err := s.pm.RetrieveProfile(snacPayloadIn.ScreenName)
		if err != nil {
			return oscar.XMessage{}, err
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

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.LOCATE,
			SubGroup:  oscar.LocateUserInfoReply,
		},
		SnacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
			TLVUserInfo: buddySess.TLVUserInfo(),
			LocateInfo: oscar.TLVRestBlock{
				TLVList: list,
			},
		},
	}, nil
}

func (s LocateService) SetDirInfoHandler(_ context.Context) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.LOCATE,
			SubGroup:  oscar.LocateSetDirReply,
		},
		SnacOut: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
			Result: 1,
		},
	}
}

func (s LocateService) SetKeywordInfoHandler(_ context.Context) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.LOCATE,
			SubGroup:  oscar.LocateSetKeywordReply,
		},
		SnacOut: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
			Unknown: 1,
		},
	}
}
