package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"log/slog"
	"time"
)

type OServiceHandler interface {
	WriteOServiceHostOnline(w io.Writer, sequence *uint32) error
	ClientOnlineHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *Session, sm SessionManager, fm FeedbagManager, room ChatRoom) error
	ClientVersionsHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x17_OServiceClientVersions) XMessage
	IdleNotificationHandler(ctx context.Context, sess *Session, sm SessionManager, fm *FeedbagStore, snacPayloadIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error
	RateParamsQueryHandler(ctx context.Context) XMessage
	RateParamsSubAddHandler(context.Context, oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd)
	ServiceRequestHandler(ctx context.Context, cfg Config, cr *ChatRegistry, sess *Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (XMessage, error)
	SetUserInfoFieldsHandler(ctx context.Context, sess *Session, sm SessionManager, fm *FeedbagStore, snacPayloadIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (XMessage, error)
	UserInfoQueryHandler(ctx context.Context, sess *Session) XMessage
}

func NewOServiceRouter(logger *slog.Logger) OServiceRouter {
	return OServiceRouter{
		OServiceHandler: OServiceService{},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type OServiceRouter struct {
	OServiceHandler
	RouteLogger
}

func (rt OServiceRouter) RouteOService(ctx context.Context, cfg Config, cr *ChatRegistry, sm SessionManager, fm *FeedbagStore, sess *Session, room ChatRoom, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.OServiceClientOnline:
		inSNAC := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user signed on")
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.ClientOnlineHandler(ctx, inSNAC, sess, sm, fm, room)
	case oscar.OServiceServiceRequest:
		inSNAC := oscar.SNAC_0x01_0x04_OServiceServiceRequest{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ServiceRequestHandler(ctx, cfg, cr, sess, inSNAC)
		switch {
		case errors.Is(err, ErrUnsupportedSubGroup):
			return sendInvalidSNACErr(SNACFrame, w, sequence)
		case err != nil:
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.OServiceRateParamsQuery:
		outSNAC := rt.RateParamsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
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
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.OServiceIdleNotification:
		inSNAC := oscar.SNAC_0x01_0x11_OServiceIdleNotification{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.IdleNotificationHandler(ctx, sess, sm, fm, inSNAC)
	case oscar.OServiceClientVersions:
		inSNAC := oscar.SNAC_0x01_0x17_OServiceClientVersions{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.ClientVersionsHandler(ctx, inSNAC)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.OServiceSetUserInfoFields:
		inSNAC := oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.SetUserInfoFieldsHandler(ctx, sess, sm, fm, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type OServiceService struct {
}

func (s OServiceService) WriteOServiceHostOnline(w io.Writer, sequence *uint32) error {
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

func (s OServiceService) ClientVersionsHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x17_OServiceClientVersions) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceHostVersions,
		},
		snacOut: oscar.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: snacPayloadIn.Versions,
		},
	}
}

func (s OServiceService) RateParamsQueryHandler(ctx context.Context) XMessage {
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

	return XMessage{
		snacFrame: snacFrameOut,
		snacOut:   snacPayloadOut,
	}
}

func (s OServiceService) UserInfoQueryHandler(ctx context.Context, sess *Session) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceUserInfoUpdate,
		},
		snacOut: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.GetTLVUserInfo(),
		},
	}
}

func (s OServiceService) ClientOnlineHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *Session, sm SessionManager, fm FeedbagManager, room ChatRoom) error {
	if err := BroadcastArrival(ctx, sess, sm, fm); err != nil {
		return err
	}
	buddies, err := fm.Buddies(sess.ScreenName)
	if err != nil {
		return err
	}
	for _, buddy := range buddies {
		err := UnicastArrival(ctx, buddy, sess.ScreenName, sm)
		switch {
		case errors.Is(err, ErrSessNotFound):
			continue
		case err != nil:
			return err
		}
	}
	return nil
}

func (s OServiceService) SetUserInfoFieldsHandler(ctx context.Context, sess *Session, sm SessionManager, fm *FeedbagStore, snacPayloadIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (XMessage, error) {
	if status, hasStatus := snacPayloadIn.GetUint32(0x06); hasStatus {
		switch status {
		case 0x000:
			sess.SetInvisible(false)
			if err := BroadcastArrival(ctx, sess, sm, fm); err != nil {
				return XMessage{}, err
			}
		case 0x100:
			sess.SetInvisible(true)
			if err := BroadcastDeparture(ctx, sess, sm, fm); err != nil {
				return XMessage{}, err
			}
		default:
			return XMessage{}, fmt.Errorf("don't know what to do with status %d", status)
		}
	}
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceUserInfoUpdate,
		},
		snacOut: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.GetTLVUserInfo(),
		},
	}, nil
}

func (s OServiceService) IdleNotificationHandler(ctx context.Context, sess *Session, sm SessionManager, fm *FeedbagStore, snacPayloadIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error {
	if snacPayloadIn.IdleTime == 0 {
		sess.SetActive()
	} else {
		sess.SetIdle(time.Duration(snacPayloadIn.IdleTime) * time.Second)
	}
	return BroadcastArrival(ctx, sess, sm, fm)
}

func (s OServiceService) ServiceRequestHandler(ctx context.Context, cfg Config, cr *ChatRegistry, sess *Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (XMessage, error) {
	if snacPayloadIn.FoodGroup != oscar.CHAT {
		return XMessage{}, ErrUnsupportedSubGroup
	}

	roomMeta, ok := snacPayloadIn.GetSlice(0x01)
	if !ok {
		return XMessage{}, errors.New("missing room info")
	}

	roomSnac := oscar.SNAC_0x01_0x04_TLVRoomInfo{}
	if err := oscar.Unmarshal(&roomSnac, bytes.NewBuffer(roomMeta)); err != nil {
		return XMessage{}, err
	}

	room, err := cr.Retrieve(string(roomSnac.Cookie))
	if err != nil {
		return XMessage{}, ErrUnsupportedSubGroup
	}
	room.NewSessionWithSN(sess.ID, sess.ScreenName)

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.OSERVICE,
			SubGroup:  oscar.OServiceServiceResponse,
		},
		snacOut: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.OServiceTLVTagsReconnectHere, Address(cfg.OSCARHost, cfg.ChatPort)),
					oscar.NewTLV(oscar.OServiceTLVTagsLoginCookie, ChatCookie{
						Cookie: []byte(room.Cookie),
						SessID: sess.ID,
					}),
					oscar.NewTLV(oscar.OServiceTLVTagsGroupID, oscar.CHAT),
					oscar.NewTLV(oscar.OServiceTLVTagsSSLCertName, ""),
					oscar.NewTLV(oscar.OServiceTLVTagsSSLState, uint8(0x00)),
				},
			},
		},
	}, nil
}

// RateParamsSubAddHandler exists to capture the SNAC input in unit tests to
// verify it's correctly unmarshalled.
func (s OServiceService) RateParamsSubAddHandler(context.Context, oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd) {
}

func NewOServiceRouterForChat(logger *slog.Logger) OServiceRouter {
	return OServiceRouter{
		OServiceHandler: OServiceServiceForChat{},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type OServiceServiceForChat struct {
	OServiceService
	RouteLogger
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

func (s OServiceServiceForChat) ClientOnlineHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *Session, sm SessionManager, fm FeedbagManager, room ChatRoom) error {
	SendChatRoomInfoUpdate(ctx, sess, sm, room)
	AlertUserJoined(ctx, sess, sm)
	SetOnlineChatUsers(ctx, sess, sm)
	return nil
}
