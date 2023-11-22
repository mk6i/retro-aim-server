package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/mkaminski/goaim/state"
)

func NewOServiceService(cfg server.Config, sm SessionManager, fm FeedbagManager) *OServiceService {
	return &OServiceService{cfg: cfg, sessionManager: sm, feedbagManager: fm}
}

type OServiceService struct {
	cfg            server.Config
	feedbagManager FeedbagManager
	sessionManager SessionManager
}

func (s OServiceService) ClientVersionsHandler(_ context.Context, snacPayloadIn oscar.SNAC_0x01_0x17_OServiceClientVersions) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceHostVersions,
		},
		Body: oscar.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: snacPayloadIn.Versions,
		},
	}
}

func (s OServiceService) RateParamsQueryHandler(_ context.Context) oscar.SNACMessage {
	snacFrameOut := oscar.SNACFrame{
		FoodGroup: oscar.OService,
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

	return oscar.SNACMessage{
		Frame: snacFrameOut,
		Body:  snacPayloadOut,
	}
}

func (s OServiceService) UserInfoQueryHandler(_ context.Context, sess *state.Session) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceUserInfoUpdate,
		},
		Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}
}

func (s OServiceService) SetUserInfoFieldsHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (oscar.SNACMessage, error) {
	if status, hasStatus := snacPayloadIn.GetUint32(0x06); hasStatus {
		switch status {
		case 0x000:
			sess.SetInvisible(false)
			if err := broadcastArrival(ctx, sess, s.sessionManager, s.feedbagManager); err != nil {
				return oscar.SNACMessage{}, err
			}
		case 0x100:
			sess.SetInvisible(true)
			if err := broadcastDeparture(ctx, sess, s.sessionManager, s.feedbagManager); err != nil {
				return oscar.SNACMessage{}, err
			}
		default:
			return oscar.SNACMessage{}, fmt.Errorf("don't know what to do with status %d", status)
		}
	}
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceUserInfoUpdate,
		},
		Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}, nil
}

func (s OServiceService) IdleNotificationHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error {
	if snacPayloadIn.IdleTime == 0 {
		sess.SetActive()
	} else {
		sess.SetIdle(time.Duration(snacPayloadIn.IdleTime) * time.Second)
	}
	return broadcastArrival(ctx, sess, s.sessionManager, s.feedbagManager)
}

// RateParamsSubAddHandler exists to capture the SNAC input in unit tests to
// verify it's correctly unmarshalled.
func (s OServiceService) RateParamsSubAddHandler(context.Context, oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd) {
}

func NewOServiceServiceForBOS(oserviceService OServiceService, cr *state.ChatRegistry) *OServiceServiceForBOS {
	return &OServiceServiceForBOS{
		OServiceService: oserviceService,
		cr:              cr,
	}
}

type OServiceServiceForBOS struct {
	OServiceService
	cr *state.ChatRegistry
}

func (s OServiceServiceForBOS) ServiceRequestHandler(_ context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (oscar.SNACMessage, error) {
	if snacPayloadIn.FoodGroup != oscar.Chat {
		return oscar.SNACMessage{}, server.ErrUnsupportedSubGroup
	}

	roomMeta, ok := snacPayloadIn.GetSlice(0x01)
	if !ok {
		return oscar.SNACMessage{}, errors.New("missing room info")
	}

	roomSnac := oscar.SNAC_0x01_0x04_TLVRoomInfo{}
	if err := oscar.Unmarshal(&roomSnac, bytes.NewBuffer(roomMeta)); err != nil {
		return oscar.SNACMessage{}, err
	}

	room, chatSessMgr, err := s.cr.Retrieve(string(roomSnac.Cookie))
	if err != nil {
		return oscar.SNACMessage{}, server.ErrUnsupportedSubGroup
	}
	chatSessMgr.(ChatSessionManager).NewSessionWithSN(sess.ID(), sess.ScreenName())

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceServiceResponse,
		},
		Body: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.OServiceTLVTagsReconnectHere, server.Address(s.cfg.OSCARHost, s.cfg.ChatPort)),
					oscar.NewTLV(oscar.OServiceTLVTagsLoginCookie, server.ChatCookie{
						Cookie: []byte(room.Cookie),
						SessID: sess.ID(),
					}),
					oscar.NewTLV(oscar.OServiceTLVTagsGroupID, oscar.Chat),
					oscar.NewTLV(oscar.OServiceTLVTagsSSLCertName, ""),
					oscar.NewTLV(oscar.OServiceTLVTagsSSLState, uint8(0x00)),
				},
			},
		},
	}, nil
}

func (s OServiceServiceForBOS) WriteOServiceHostOnline() oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceHostOnline,
		},
		Body: oscar.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				oscar.Alert,
				oscar.Buddy,
				oscar.ChatNav,
				oscar.Feedbag,
				oscar.ICBM,
				oscar.Locate,
				oscar.OService,
			},
		},
	}
}

func (s OServiceServiceForBOS) ClientOnlineHandler(ctx context.Context, _ oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	if err := broadcastArrival(ctx, sess, s.sessionManager, s.feedbagManager); err != nil {
		return err
	}
	buddies, err := s.feedbagManager.Buddies(sess.ScreenName())
	if err != nil {
		return err
	}
	for _, buddy := range buddies {
		unicastArrival(ctx, buddy, sess.ScreenName(), s.sessionManager)
	}
	return nil
}

func NewOServiceServiceForChat(oserviceService OServiceService, chatRegistry *state.ChatRegistry) *OServiceServiceForChat {
	return &OServiceServiceForChat{
		OServiceService: oserviceService,
		chatRegistry:    chatRegistry,
	}
}

type OServiceServiceForChat struct {
	OServiceService
	chatRegistry *state.ChatRegistry
}

func (s OServiceServiceForChat) WriteOServiceHostOnline() oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceHostOnline,
		},
		Body: oscar.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{oscar.OService, oscar.Chat},
		},
	}
}

func (s OServiceServiceForChat) ClientOnlineHandler(ctx context.Context, _ oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session, chatID string) error {
	room, chatSessMgr, err := s.chatRegistry.Retrieve(chatID)
	if err != nil {
		return err
	}
	sendChatRoomInfoUpdate(ctx, sess, chatSessMgr.(ChatSessionManager), room)
	alertUserJoined(ctx, sess, chatSessMgr.(ChatSessionManager))
	setOnlineChatUsers(ctx, sess, chatSessMgr.(ChatSessionManager))
	return nil
}
