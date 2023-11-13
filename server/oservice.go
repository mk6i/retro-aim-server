package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/user"
	"io"
	"log/slog"
	"time"
)

type OServiceHandler interface {
	ClientVersionsHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x17_OServiceClientVersions) oscar.XMessage
	IdleNotificationHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error
	RateParamsQueryHandler(ctx context.Context) oscar.XMessage
	RateParamsSubAddHandler(context.Context, oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd)
	SetUserInfoFieldsHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (oscar.XMessage, error)
	UserInfoQueryHandler(ctx context.Context, sess *user.Session) oscar.XMessage
}

type OServiceBOSHandler interface {
	OServiceHandler
	WriteOServiceHostOnline(w io.Writer, sequence *uint32) error
	ServiceRequestHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (oscar.XMessage, error)
	ClientOnlineHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *user.Session) error
}

type OServiceChatHandler interface {
	OServiceHandler
	WriteOServiceHostOnline(w io.Writer, sequence *uint32) error
	ServiceRequestHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (oscar.XMessage, error)
	ClientOnlineHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *user.Session, room ChatRoom) error
}

type OServiceRouter struct {
	OServiceHandler
	RouteLogger
}

func (rt OServiceRouter) RouteOService(ctx context.Context, sess *user.Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.OServiceRateParamsQuery:
		outSNAC := rt.RateParamsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceRateParamsSubAdd:
		inSNAC := oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.RateParamsSubAddHandler(ctx, inSNAC)
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return oscar.Unmarshal(&inSNAC, r)
	case oscar.OServiceUserInfoQuery:
		outSNAC := rt.UserInfoQueryHandler(ctx, sess)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceIdleNotification:
		inSNAC := oscar.SNAC_0x01_0x11_OServiceIdleNotification{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.IdleNotificationHandler(ctx, sess, inSNAC)
	case oscar.OServiceClientVersions:
		inSNAC := oscar.SNAC_0x01_0x17_OServiceClientVersions{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.ClientVersionsHandler(ctx, inSNAC)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceSetUserInfoFields:
		inSNAC := oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.SetUserInfoFieldsHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

func NewOServiceRouterForBOS(logger *slog.Logger, cfg Config, fm FeedbagManager, sm SessionManager, cr *ChatRegistry) OServiceBOSRouter {
	oss := OServiceService{
		cfg: cfg,
		fm:  fm,
		sm:  sm,
	}
	return OServiceBOSRouter{
		OServiceRouter: OServiceRouter{
			OServiceHandler: oss,
			RouteLogger: RouteLogger{
				Logger: logger,
			},
		},
		OServiceBOSHandler: OServiceServiceForBOS{
			OServiceService: oss,
			cr:              cr,
		},
	}
}

type OServiceBOSRouter struct {
	OServiceRouter
	OServiceBOSHandler
}

func (rt OServiceBOSRouter) RouteOService(ctx context.Context, sess *user.Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.OServiceServiceRequest:
		inSNAC := oscar.SNAC_0x01_0x04_OServiceServiceRequest{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ServiceRequestHandler(ctx, sess, inSNAC)
		switch {
		case errors.Is(err, ErrUnsupportedSubGroup):
			return sendInvalidSNACErr(SNACFrame, w, sequence)
		case err != nil:
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceClientOnline:
		inSNAC := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user signed on")
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.OServiceBOSHandler.ClientOnlineHandler(ctx, inSNAC, sess)
	default:
		return rt.OServiceRouter.RouteOService(ctx, sess, SNACFrame, r, w, sequence)
	}
}

type OServiceChatRouter struct {
	OServiceRouter
	OServiceChatHandler
}

func (rt OServiceChatRouter) RouteOService(ctx context.Context, sess *user.Session, room ChatRoom, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.OServiceServiceRequest:
		inSNAC := oscar.SNAC_0x01_0x04_OServiceServiceRequest{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ServiceRequestHandler(ctx, sess, inSNAC)
		switch {
		case errors.Is(err, ErrUnsupportedSubGroup):
			return sendInvalidSNACErr(SNACFrame, w, sequence)
		case err != nil:
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceClientOnline:
		inSNAC := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user signed on")
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.OServiceChatHandler.ClientOnlineHandler(ctx, inSNAC, sess, room)
	default:
		return rt.OServiceRouter.RouteOService(ctx, sess, SNACFrame, r, w, sequence)
	}
}

type OServiceService struct {
	cfg Config
	fm  FeedbagManager
	sm  SessionManager
}

func (s OServiceService) ClientVersionsHandler(_ context.Context, snacPayloadIn oscar.SNAC_0x01_0x17_OServiceClientVersions) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceHostVersions,
		},
		SnacOut: oscar.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: snacPayloadIn.Versions,
		},
	}
}

func (s OServiceService) RateParamsQueryHandler(_ context.Context) oscar.XMessage {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: oscar.OSERVICE,
		SubGroup:  oscar.OServiceRateParamsReply,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x07_OServiceRateParamsReply{
		RateClasses: []struct {
			ID              uint16
			WindowSize      uint32
			ClearLevel      uint32
			AlertLevel      uint32
			LimitLevel      uint32
			DisconnectLevel uint32
			CurrentLevel    uint32
			MaxLevel        uint32
			LastTime        uint32 // protocol v2 only
			CurrentState    uint8  // protocol v2 only
		}{
			{
				ID:              0x0001,
				WindowSize:      0x00000050,
				ClearLevel:      0x000009C4,
				AlertLevel:      0x000007D0,
				LimitLevel:      0x000005DC,
				DisconnectLevel: 0x00000320,
				CurrentLevel:    0x00000D69,
				MaxLevel:        0x00001770,
				LastTime:        0x00000000,
				CurrentState:    0x00,
			},
		},
		RateGroups: []struct {
			ID    uint16
			Pairs []struct {
				FoodGroup uint16
				SubGroup  uint16
			} `count_prefix:"uint16"`
		}{
			{
				ID: 1,
				Pairs: []struct {
					FoodGroup uint16
					SubGroup  uint16
				}{},
			},
		},
	}

	for i := uint16(0); i < 24; i++ { // for each food group
		for j := uint16(0); j < 32; j++ { // for each subgroup
			snacPayloadOut.RateGroups[0].Pairs = append(snacPayloadOut.RateGroups[0].Pairs,
				struct {
					FoodGroup uint16
					SubGroup  uint16
				}{
					FoodGroup: i,
					SubGroup:  j,
				})
		}
	}

	return oscar.XMessage{
		SnacFrame: snacFrameOut,
		SnacOut:   snacPayloadOut,
	}
}

func (s OServiceService) UserInfoQueryHandler(_ context.Context, sess *user.Session) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceUserInfoUpdate,
		},
		SnacOut: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}
}

func (s OServiceService) SetUserInfoFieldsHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (oscar.XMessage, error) {
	if status, hasStatus := snacPayloadIn.GetUint32(0x06); hasStatus {
		switch status {
		case 0x000:
			sess.SetInvisible(false)
			if err := BroadcastArrival(ctx, sess, s.sm, s.fm); err != nil {
				return oscar.XMessage{}, err
			}
		case 0x100:
			sess.SetInvisible(true)
			if err := BroadcastDeparture(ctx, sess, s.sm, s.fm); err != nil {
				return oscar.XMessage{}, err
			}
		default:
			return oscar.XMessage{}, fmt.Errorf("don't know what to do with status %d", status)
		}
	}
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceUserInfoUpdate,
		},
		SnacOut: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}, nil
}

func (s OServiceService) IdleNotificationHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error {
	if snacPayloadIn.IdleTime == 0 {
		sess.SetActive()
	} else {
		sess.SetIdle(time.Duration(snacPayloadIn.IdleTime) * time.Second)
	}
	return BroadcastArrival(ctx, sess, s.sm, s.fm)
}

// RateParamsSubAddHandler exists to capture the SNAC input in unit tests to
// verify it's correctly unmarshalled.
func (s OServiceService) RateParamsSubAddHandler(context.Context, oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd) {
}

type OServiceServiceForBOS struct {
	OServiceService
	cr *ChatRegistry
}

func (s OServiceServiceForBOS) ServiceRequestHandler(_ context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (oscar.XMessage, error) {
	if snacPayloadIn.FoodGroup != oscar.CHAT {
		return oscar.XMessage{}, ErrUnsupportedSubGroup
	}

	roomMeta, ok := snacPayloadIn.GetSlice(0x01)
	if !ok {
		return oscar.XMessage{}, errors.New("missing room info")
	}

	roomSnac := oscar.SNAC_0x01_0x04_TLVRoomInfo{}
	if err := oscar.Unmarshal(&roomSnac, bytes.NewBuffer(roomMeta)); err != nil {
		return oscar.XMessage{}, err
	}

	room, err := s.cr.Retrieve(string(roomSnac.Cookie))
	if err != nil {
		return oscar.XMessage{}, ErrUnsupportedSubGroup
	}
	room.NewSessionWithSN(sess.ID(), sess.ScreenName())

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceServiceResponse,
		},
		SnacOut: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.OServiceTLVTagsReconnectHere, Address(s.cfg.OSCARHost, s.cfg.ChatPort)),
					oscar.NewTLV(oscar.OServiceTLVTagsLoginCookie, ChatCookie{
						Cookie: []byte(room.Cookie),
						SessID: sess.ID(),
					}),
					oscar.NewTLV(oscar.OServiceTLVTagsGroupID, oscar.CHAT),
					oscar.NewTLV(oscar.OServiceTLVTagsSSLCertName, ""),
					oscar.NewTLV(oscar.OServiceTLVTagsSSLState, uint8(0x00)),
				},
			},
		},
	}, nil
}

func (s OServiceServiceForBOS) WriteOServiceHostOnline(w io.Writer, sequence *uint32) error {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: oscar.OSERVICE,
		SubGroup:  oscar.OServiceHostOnline,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x03_OServiceHostOnline{
		FoodGroups: []uint16{
			oscar.ALERT,
			oscar.BUDDY,
			oscar.CHAT_NAV,
			oscar.FEEDBAG,
			oscar.ICBM,
			oscar.LOCATE,
			oscar.OSERVICE,
		},
	}
	return writeOutSNAC(oscar.SnacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}

func (s OServiceServiceForBOS) ClientOnlineHandler(ctx context.Context, _ oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *user.Session) error {
	if err := BroadcastArrival(ctx, sess, s.sm, s.fm); err != nil {
		return err
	}
	buddies, err := s.fm.Buddies(sess.ScreenName())
	if err != nil {
		return err
	}
	for _, buddy := range buddies {
		err := UnicastArrival(ctx, buddy, sess.ScreenName(), s.sm)
		switch {
		case errors.Is(err, ErrSessNotFound):
			continue
		case err != nil:
			return err
		}
	}
	return nil
}

func NewOServiceRouterForChat(logger *slog.Logger, cfg Config, fm FeedbagManager, sm SessionManager) OServiceChatRouter {
	oss := OServiceService{
		cfg: cfg,
		fm:  fm,
		sm:  sm,
	}
	return OServiceChatRouter{
		OServiceRouter: OServiceRouter{
			OServiceHandler: oss,
			RouteLogger: RouteLogger{
				Logger: logger,
			},
		},
		OServiceChatHandler: OServiceServiceForChat{
			OServiceService: oss,
		},
	}
}

type OServiceServiceForChat struct {
	OServiceService
	RouteLogger
}

func (s OServiceServiceForChat) ServiceRequestHandler(_ context.Context, _ *user.Session, _ oscar.SNAC_0x01_0x04_OServiceServiceRequest) (oscar.XMessage, error) {
	return oscar.XMessage{}, ErrUnsupportedSubGroup
}

func (s OServiceServiceForChat) WriteOServiceHostOnline(w io.Writer, sequence *uint32) error {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: oscar.OSERVICE,
		SubGroup:  oscar.OServiceHostOnline,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x03_OServiceHostOnline{
		FoodGroups: []uint16{oscar.OSERVICE, oscar.CHAT},
	}
	return writeOutSNAC(oscar.SnacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}

func (s OServiceServiceForChat) ClientOnlineHandler(ctx context.Context, _ oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *user.Session, room ChatRoom) error {
	SendChatRoomInfoUpdate(ctx, sess, room)
	AlertUserJoined(ctx, sess, room)
	SetOnlineChatUsers(ctx, sess, room)
	return nil
}
