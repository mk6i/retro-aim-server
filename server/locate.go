package server

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
)

type LocateHandler interface {
	RightsQueryHandler(ctx context.Context) XMessage
	SetDirInfoHandler(ctx context.Context) XMessage
	SetInfoHandler(ctx context.Context, sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error
	SetKeywordInfoHandler(ctx context.Context) XMessage
	UserInfoQuery2Handler(ctx context.Context, sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (XMessage, error)
}

func NewLocateRouter(logger *slog.Logger) LocateRouter {
	return LocateRouter{
		LocateHandler: LocateService{},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type LocateRouter struct {
	LocateHandler
	RouteLogger
}

func (rt LocateRouter) RouteLocate(ctx context.Context, sess *Session, sm SessionManager, fm *FeedbagStore, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.LocateRightsQuery:
		outSNAC := rt.RightsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.LocateSetInfo:
		inSNAC := oscar.SNAC_0x02_0x04_LocateSetInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.SetInfoHandler(ctx, sess, sm, fm, fm, inSNAC)
	case oscar.LocateSetDirInfo:
		inSNAC := oscar.SNAC_0x02_0x09_LocateSetDirInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.SetDirInfoHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
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
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.LocateUserInfoQuery2:
		inSNAC := oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.UserInfoQuery2Handler(ctx, sess, sm, fm, fm, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type LocateService struct {
}

func (s LocateService) RightsQueryHandler(context.Context) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.LOCATE,
			SubGroup:  oscar.LocateRightsReply,
		},
		snacOut: oscar.SNAC_0x02_0x03_LocateRightsReply{
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

func (s LocateService) SetInfoHandler(ctx context.Context, sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snacPayloadIn oscar.SNAC_0x02_0x04_LocateSetInfo) error {
	// update profile
	if profile, hasProfile := snacPayloadIn.GetString(oscar.LocateTLVTagsInfoSigData); hasProfile {
		if err := pm.UpsertProfile(sess.ScreenName, profile); err != nil {
			return err
		}
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := snacPayloadIn.GetString(oscar.LocateTLVTagsInfoUnavailableData); hasAwayMsg {
		sess.SetAwayMessage(awayMsg)
		if err := BroadcastArrival(ctx, sess, sm, fm); err != nil {
			return err
		}
	}
	return nil
}

func (s LocateService) UserInfoQuery2Handler(ctx context.Context, sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snacPayloadIn oscar.SNAC_0x02_0x15_LocateUserInfoQuery2) (XMessage, error) {
	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)
	switch {
	case err != nil:
		return XMessage{}, err
	case blocked != BlockedNo:
		return XMessage{
			snacFrame: oscar.SnacFrame{
				FoodGroup: oscar.LOCATE,
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
				FoodGroup: oscar.LOCATE,
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
			oscar.NewTLV(oscar.LocateTLVTagsInfoSigMime, `text/aolrtf; charset="us-ascii"`),
			oscar.NewTLV(oscar.LocateTLVTagsInfoSigData, profile),
		})
	}

	if snacPayloadIn.RequestAwayMessage() {
		list.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableMime, `text/aolrtf; charset="us-ascii"`),
			oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableData, buddySess.GetAwayMessage()),
		})
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.LOCATE,
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

func (s LocateService) SetDirInfoHandler(ctx context.Context) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.LOCATE,
			SubGroup:  oscar.LocateSetDirReply,
		},
		snacOut: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
			Result: 1,
		},
	}
}

func (s LocateService) SetKeywordInfoHandler(ctx context.Context) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.LOCATE,
			SubGroup:  oscar.LocateSetKeywordReply,
		},
		snacOut: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
			Unknown: 1,
		},
	}
}
