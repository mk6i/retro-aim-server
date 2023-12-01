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

// NewOServiceService creates a new instance of OServiceService.
func NewOServiceService(cfg server.Config, messageRelayer MessageRelayer, feedbagManager FeedbagManager) *OServiceService {
	return &OServiceService{cfg: cfg, messageRelayer: messageRelayer, feedbagManager: feedbagManager}
}

// OServiceService contains handlers for the OService food group.
type OServiceService struct {
	cfg            server.Config
	feedbagManager FeedbagManager
	messageRelayer MessageRelayer
}

// ClientVersionsHandler informs the server what food group versions the client
// supports and returns to the client what food group versions it supports.
// This method simply regurgitates versions supplied by the client in inBody
// back to the client in a OServiceHostVersions SNAC. The server doesn't
// attempt to accommodate any particular food group version. The server
// implicitly accommodates any food group version for Windows AIM clients 5.x.
// It returns SNAC oscar.OServiceHostVersions containing the server's supported
// food group versions.
func (s OServiceService) ClientVersionsHandler(_ context.Context, frame oscar.SNACFrame, inBody oscar.SNAC_0x01_0x17_OServiceClientVersions) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceHostVersions,
			RequestID: frame.RequestID,
		},
		Body: oscar.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: inBody.Versions,
		},
	}
}

// RateParamsQueryHandler returns SNAC rate limits.
// The purpose of this method is to provide information about rate limits that
// can be enforced on the server side. The response consists of two main parts:
// rate classes and rate groups. Rate classes define limits based on specific
// parameters, while rate groups associate these limits with relevant SNAC
// types.
//
// Note: The current implementation does not enforce server-side rate limiting.
// Instead, the provided values inform the client about the recommended
// client-side rate limits.
//
// It returns SNAC osca.rOServiceRateParamsReply containing rate limits for
// sending Instant Messages (IMs) and chat messages. More refined limits may be
// added in the future if/when server rate limiting is implemented.
func (s OServiceService) RateParamsQueryHandler(_ context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceRateParamsReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x01_0x07_OServiceRateParamsReply{
			RateClasses: []struct {
				ID              uint16
				WindowSize      uint32
				ClearLevel      uint32
				AlertLevel      uint32
				LimitLevel      uint32
				DisconnectLevel uint32
				CurrentLevel    uint32
				MaxLevel        uint32
				LastTime        uint32
				CurrentState    uint8
			}{
				{
					ID:              0x01,
					WindowSize:      0x0050,
					ClearLevel:      0x09C4,
					AlertLevel:      0x07D0,
					LimitLevel:      0x05DC,
					DisconnectLevel: 0x0320,
					CurrentLevel:    0x0D69,
					MaxLevel:        0x1770,
					LastTime:        0x0000,
					CurrentState:    0x0,
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
					}{
						{
							FoodGroup: oscar.Buddy,
							SubGroup:  oscar.BuddyRightsQuery,
						},
						{
							FoodGroup: oscar.Chat,
							SubGroup:  oscar.ChatChannelMsgToHost,
						},
						{
							FoodGroup: oscar.ChatNav,
							SubGroup:  oscar.ChatNavRequestChatRights,
						},
						{
							FoodGroup: oscar.ChatNav,
							SubGroup:  oscar.ChatNavRequestRoomInfo,
						},
						{
							FoodGroup: oscar.ChatNav,
							SubGroup:  oscar.ChatNavCreateRoom,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagRightsQuery,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagQuery,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagQueryIfModified,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagUse,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagInsertItem,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagUpdateItem,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagDeleteItem,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagStartCluster,
						},
						{
							FoodGroup: oscar.Feedbag,
							SubGroup:  oscar.FeedbagEndCluster,
						},
						{
							FoodGroup: oscar.ICBM,
							SubGroup:  oscar.ICBMAddParameters,
						},
						{
							FoodGroup: oscar.ICBM,
							SubGroup:  oscar.ICBMParameterQuery,
						},
						{
							FoodGroup: oscar.ICBM,
							SubGroup:  oscar.ICBMChannelMsgToHost,
						},
						{
							FoodGroup: oscar.ICBM,
							SubGroup:  oscar.ICBMEvilRequest,
						},
						{
							FoodGroup: oscar.ICBM,
							SubGroup:  oscar.ICBMClientErr,
						},
						{
							FoodGroup: oscar.ICBM,
							SubGroup:  oscar.ICBMClientEvent,
						},
						{
							FoodGroup: oscar.Locate,
							SubGroup:  oscar.LocateRightsQuery,
						},
						{
							FoodGroup: oscar.Locate,
							SubGroup:  oscar.LocateSetInfo,
						},
						{
							FoodGroup: oscar.Locate,
							SubGroup:  oscar.LocateSetDirInfo,
						},
						{
							FoodGroup: oscar.Locate,
							SubGroup:  oscar.LocateGetDirInfo,
						},
						{
							FoodGroup: oscar.Locate,
							SubGroup:  oscar.LocateSetKeywordInfo,
						},
						{
							FoodGroup: oscar.Locate,
							SubGroup:  oscar.LocateUserInfoQuery2,
						},
						{
							FoodGroup: oscar.OService,
							SubGroup:  oscar.OServiceServiceRequest,
						},
						{
							FoodGroup: oscar.OService,
							SubGroup:  oscar.OServiceClientOnline,
						},
						{
							FoodGroup: oscar.OService,
							SubGroup:  oscar.OServiceRateParamsQuery,
						},
						{
							FoodGroup: oscar.OService,
							SubGroup:  oscar.OServiceRateParamsSubAdd,
						},
						{
							FoodGroup: oscar.OService,
							SubGroup:  oscar.OServiceUserInfoQuery,
						},
						{
							FoodGroup: oscar.OService,
							SubGroup:  oscar.OServiceIdleNotification,
						},
						{
							FoodGroup: oscar.OService,
							SubGroup:  oscar.OServiceClientVersions,
						},
						{
							FoodGroup: oscar.OService,
							SubGroup:  oscar.OServiceSetUserInfoFields,
						},
					},
				},
			},
		},
	}
}

// UserInfoQueryHandler returns SNAC oscar.OServiceUserInfoUpdate containing
// the user's info.
func (s OServiceService) UserInfoQueryHandler(_ context.Context, sess *state.Session, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceUserInfoUpdate,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}
}

// SetUserInfoFieldsHandler sets the user's visibility status to visible or invisible.
// The visibility status is set according to the inFrame TLV entry under key
// oscar.OServiceUserInfoStatus. If the value is 0x0000, set invisible. If set
// to 0x0100, set invisible. Else, return an error for any other value.
// It returns SNAC oscar.OServiceUserInfoUpdate containing the user's info.
func (s OServiceService) SetUserInfoFieldsHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (oscar.SNACMessage, error) {
	if status, hasStatus := inBody.GetUint32(oscar.OServiceUserInfoStatus); hasStatus {
		switch status {
		case 0x0000:
			sess.SetInvisible(false)
			if err := broadcastArrival(ctx, sess, s.messageRelayer, s.feedbagManager); err != nil {
				return oscar.SNACMessage{}, err
			}
		case 0x0100:
			sess.SetInvisible(true)
			if err := broadcastDeparture(ctx, sess, s.messageRelayer, s.feedbagManager); err != nil {
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
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}, nil
}

// IdleNotificationHandler sets the user idle time.
// Set session idle time to the value of bodyIn.IdleTime. Return a user arrival
// message to all users who have this user on their buddy list.
func (s OServiceService) IdleNotificationHandler(ctx context.Context, sess *state.Session, bodyIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error {
	if bodyIn.IdleTime == 0 {
		sess.SetActive()
	} else {
		sess.SetIdle(time.Duration(bodyIn.IdleTime) * time.Second)
	}
	return broadcastArrival(ctx, sess, s.messageRelayer, s.feedbagManager)
}

// RateParamsSubAddHandler exists to capture the SNAC input in unit tests to
// verify it's correctly unmarshalled.
func (s OServiceService) RateParamsSubAddHandler(context.Context, oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd) {
}

// NewOServiceServiceForBOS creates a new instance of OServiceServiceForBOS.
func NewOServiceServiceForBOS(oserviceService OServiceService, cr *state.ChatRegistry) *OServiceServiceForBOS {
	return &OServiceServiceForBOS{
		OServiceService: oserviceService,
		chatRegistry:    cr,
	}
}

// OServiceServiceForBOS contains handlers for the OService food group for the
// BOS service.
type OServiceServiceForBOS struct {
	OServiceService
	chatRegistry *state.ChatRegistry
}

// ServiceRequestHandler configures food group settings for the current user.
// This method only provides services for the Chat food group; return
// server.ErrUnsupportedSubGroup for any other food group. When the chat food
// group is specified in inFrame, add user to the chat room specified by TLV
// 0x01.
// It returns SNAC oscar.OServiceServiceResponse containing metadata the client
// needs to connect to the chat service and join the chat room.
func (s OServiceServiceForBOS) ServiceRequestHandler(_ context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x01_0x04_OServiceServiceRequest) (oscar.SNACMessage, error) {
	if inBody.FoodGroup != oscar.Chat {
		return oscar.SNACMessage{}, server.ErrUnsupportedSubGroup
	}

	roomMeta, ok := inBody.GetSlice(0x01)
	if !ok {
		return oscar.SNACMessage{}, errors.New("missing room info")
	}

	roomSnac := oscar.SNAC_0x01_0x04_TLVRoomInfo{}
	if err := oscar.Unmarshal(&roomSnac, bytes.NewBuffer(roomMeta)); err != nil {
		return oscar.SNACMessage{}, err
	}

	room, chatSessMgr, err := s.chatRegistry.Retrieve(string(roomSnac.Cookie))
	if err != nil {
		return oscar.SNACMessage{}, server.ErrUnsupportedSubGroup
	}
	chatSessMgr.(SessionManager).NewSessionWithSN(sess.ID(), sess.ScreenName())

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceServiceResponse,
			RequestID: inFrame.RequestID,
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

// WriteOServiceHostOnline initiates the BOS protocol sequence.
// It returns SNAC oscar.OServiceHostOnline containing the list food groups
// supported by the BOS service.
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

// ClientOnlineHandler runs when the current user is ready to join.
// It performs the following sequence of actions:
//   - Announce current user's arrival to users who have the current user on
//     their buddy list.
//   - Send current user its buddy list
func (s OServiceServiceForBOS) ClientOnlineHandler(ctx context.Context, _ oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	if err := broadcastArrival(ctx, sess, s.messageRelayer, s.feedbagManager); err != nil {
		return err
	}
	buddies, err := s.feedbagManager.Buddies(sess.ScreenName())
	if err != nil {
		return err
	}
	for _, buddy := range buddies {
		unicastArrival(ctx, buddy, sess.ScreenName(), s.messageRelayer)
	}
	return nil
}

// NewOServiceServiceForChat creates a new instance of OServiceServiceForChat.
func NewOServiceServiceForChat(oserviceService OServiceService, chatRegistry *state.ChatRegistry) *OServiceServiceForChat {
	return &OServiceServiceForChat{
		OServiceService: oserviceService,
		chatRegistry:    chatRegistry,
	}
}

// OServiceServiceForChat contains handlers for the OService food group for the
// Chat service.
type OServiceServiceForChat struct {
	OServiceService
	chatRegistry *state.ChatRegistry
}

// WriteOServiceHostOnline initiates the Chat protocol sequence.
// It returns SNAC oscar.OServiceHostOnline containing the list of food groups
// supported by the Chat service.
func (s OServiceServiceForChat) WriteOServiceHostOnline() oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceHostOnline,
		},
		Body: oscar.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				oscar.OService,
				oscar.Chat,
			},
		},
	}
}

// ClientOnlineHandler runs when the current user is ready to join the chat.
// Trigger the following actions:
//   - Send current user the chat room metadata
//   - Announce current user's arrival to other chat room participants
//   - Send current user the chat room participant list
func (s OServiceServiceForChat) ClientOnlineHandler(ctx context.Context, sess *state.Session, chatID string) error {
	room, chatSessMgr, err := s.chatRegistry.Retrieve(chatID)
	if err != nil {
		return err
	}
	sendChatRoomInfoUpdate(ctx, sess, chatSessMgr.(ChatMessageRelayer), room)
	alertUserJoined(ctx, sess, chatSessMgr.(ChatMessageRelayer))
	setOnlineChatUsers(ctx, sess, chatSessMgr.(ChatMessageRelayer))
	return nil
}
