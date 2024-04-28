package foodgroup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewOServiceService creates a new instance of OServiceService.
func NewOServiceService(
	cfg config.Config,
	messageRelayer MessageRelayer,
	feedbagManager FeedbagManager,
	legacyBuddyListManager LegacyBuddyListManager,
) *OServiceService {
	return &OServiceService{
		cfg:                    cfg,
		feedbagManager:         feedbagManager,
		legacyBuddyListManager: legacyBuddyListManager,
		messageRelayer:         messageRelayer,
	}
}

// OServiceService provides functionality for the OService food group, which
// provides an assortment of services useful across multiple food groups.
type OServiceService struct {
	cfg                    config.Config
	feedbagManager         FeedbagManager
	legacyBuddyListManager LegacyBuddyListManager
	messageRelayer         MessageRelayer
}

// ClientVersions informs the server what food group versions the client
// supports and returns to the client what food group versions it supports.
// This method simply regurgitates versions supplied by the client in inBody
// back to the client in a OServiceHostVersions SNAC. The server doesn't
// attempt to accommodate any particular food group version. The server
// implicitly accommodates any food group version for Windows AIM clients 5.x.
// It returns SNAC wire.OServiceHostVersions containing the server's supported
// food group versions.
func (s OServiceService) ClientVersions(_ context.Context, frame wire.SNACFrame, inBody wire.SNAC_0x01_0x17_OServiceClientVersions) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostVersions,
			RequestID: frame.RequestID,
		},
		Body: wire.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: inBody.Versions,
		},
	}
}

// RateParamsQuery returns SNAC rate limits.
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
// It returns SNAC wire.OServiceRateParamsReply containing rate limits for
// sending Instant Messages (IMs) and chat messages. More refined limits may be
// added in the future if/when server rate limiting is implemented.
func (s OServiceService) RateParamsQuery(_ context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceRateParamsReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x01_0x07_OServiceRateParamsReply{
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
				// these values were taken from the example SNAC dump documented here:
				// https://web.archive.org/web/20221207225518/https://wiki.nina.chat/wiki/Protocols/OSCAR/SNAC/OSERVICE_RATE_PARAMS_REPLY
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
							FoodGroup: wire.Buddy,
							SubGroup:  wire.BuddyRightsQuery,
						},
						{
							FoodGroup: wire.Chat,
							SubGroup:  wire.ChatChannelMsgToHost,
						},
						{
							FoodGroup: wire.ChatNav,
							SubGroup:  wire.ChatNavRequestChatRights,
						},
						{
							FoodGroup: wire.ChatNav,
							SubGroup:  wire.ChatNavRequestRoomInfo,
						},
						{
							FoodGroup: wire.ChatNav,
							SubGroup:  wire.ChatNavCreateRoom,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagRightsQuery,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagQuery,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagQueryIfModified,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagUse,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagInsertItem,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagUpdateItem,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagDeleteItem,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagStartCluster,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagEndCluster,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMAddParameters,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMParameterQuery,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMChannelMsgToHost,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMEvilRequest,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMClientErr,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMClientEvent,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateRightsQuery,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateSetInfo,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateSetDirInfo,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateGetDirInfo,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateSetKeywordInfo,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateUserInfoQuery2,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceServiceRequest,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceClientOnline,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceRateParamsQuery,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceRateParamsSubAdd,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceUserInfoQuery,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceIdleNotification,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceClientVersions,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceSetUserInfoFields,
						},
						{
							FoodGroup: wire.BART,
							SubGroup:  wire.BARTUploadQuery,
						},
						{
							FoodGroup: wire.BART,
							SubGroup:  wire.BARTDownloadQuery,
						},
					},
				},
			},
		},
	}
}

// UserInfoQuery returns SNAC wire.OServiceUserInfoUpdate containing
// the user's info.
func (s OServiceService) UserInfoQuery(_ context.Context, sess *state.Session, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoUpdate,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}
}

// SetUserInfoFields sets the user's visibility status to visible or invisible.
// The visibility status is set according to the inFrame TLV entry under key
// wire.OServiceUserInfoStatus. If the value is 0x0000, set invisible. If set
// to 0x0100, set invisible. Else, return an error for any other value.
// It returns SNAC wire.OServiceUserInfoUpdate containing the user's info.
func (s OServiceService) SetUserInfoFields(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (wire.SNACMessage, error) {
	if status, hasStatus := inBody.Uint32(wire.OServiceUserInfoStatus); hasStatus {
		switch status {
		case 0x0000:
			sess.SetInvisible(false)
			if err := broadcastArrival(ctx, sess, s.messageRelayer, s.feedbagManager, s.legacyBuddyListManager); err != nil {
				return wire.SNACMessage{}, err
			}
		case 0x0100:
			sess.SetInvisible(true)
			if err := broadcastDeparture(ctx, sess, s.messageRelayer, s.feedbagManager, s.legacyBuddyListManager); err != nil {
				return wire.SNACMessage{}, err
			}
		default:
			return wire.SNACMessage{}, fmt.Errorf("don't know what to do with status %d", status)
		}
	}
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoUpdate,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}, nil
}

// IdleNotification sets the user idle time.
// Set session idle time to the value of bodyIn.IdleTime. Return a user arrival
// message to all users who have this user on their buddy list.
func (s OServiceService) IdleNotification(ctx context.Context, sess *state.Session, bodyIn wire.SNAC_0x01_0x11_OServiceIdleNotification) error {
	if bodyIn.IdleTime == 0 {
		sess.UnsetIdle()
	} else {
		sess.SetIdle(time.Duration(bodyIn.IdleTime) * time.Second)
	}
	return broadcastArrival(ctx, sess, s.messageRelayer, s.feedbagManager, s.legacyBuddyListManager)
}

// RateParamsSubAdd exists to capture the SNAC input in unit tests to
// verify it's correctly unmarshalled.
func (s OServiceService) RateParamsSubAdd(context.Context, wire.SNAC_0x01_0x08_OServiceRateParamsSubAdd) {
}

// NewOServiceServiceForBOS creates a new instance of OServiceServiceForBOS.
func NewOServiceServiceForBOS(oserviceService OServiceService, cr *state.ChatRegistry) *OServiceServiceForBOS {
	return &OServiceServiceForBOS{
		OServiceService: oserviceService,
		chatRegistry:    cr,
	}
}

// OServiceServiceForBOS provides functionality for the OService food group
// running on the BOS server.
type OServiceServiceForBOS struct {
	OServiceService
	chatRegistry *state.ChatRegistry
}

// chatLoginCookie represents credentials used to authenticate a user chat
// session.
type chatLoginCookie struct {
	Cookie string `len_prefix:"uint8"`
	SessID string `len_prefix:"uint16"`
}

// ServiceRequest handles service discovery, providing a host name and metadata
// for connecting to the food group service specified in inFrame.
// Depending on the food group specified, the method behaves as follows:
// - ChatNav: Directs the user back to the current BOS server, which provides
// ChatNav services. AIM 4.8 requests ChatNav service info even though
// HostOnline reports it as available via BOS.
// - Chat: Directs the client to the chat server along with metadata for
// connecting to the chat room specified in inFrame.
// - Other Food Groups: Returns wire.ErrUnsupportedFoodGroup.
func (s OServiceServiceForBOS) ServiceRequest(_ context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x01_0x04_OServiceServiceRequest) (wire.SNACMessage, error) {
	switch inBody.FoodGroup {
	case wire.ChatNav:
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceTLVTagsReconnectHere, config.Address(s.cfg.OSCARHost, s.cfg.ChatNavPort)),
						wire.NewTLV(wire.OServiceTLVTagsLoginCookie, sess.ID()),
						wire.NewTLV(wire.OServiceTLVTagsGroupID, wire.ChatNav),
						wire.NewTLV(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLV(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	case wire.Chat:
		roomMeta, ok := inBody.Slice(0x01)
		if !ok {
			return wire.SNACMessage{}, errors.New("missing room info")
		}

		roomSNAC := wire.SNAC_0x01_0x04_TLVRoomInfo{}
		if err := wire.Unmarshal(&roomSNAC, bytes.NewBuffer(roomMeta)); err != nil {
			return wire.SNACMessage{}, err
		}

		room, chatSessMgr, err := s.chatRegistry.Retrieve(roomSNAC.Cookie)
		if err != nil {
			return wire.SNACMessage{}, err
		}
		chatSess := chatSessMgr.(SessionManager).AddSession(sess.ID(), sess.ScreenName())
		chatSess.SetChatRoomCookie(room.Cookie)

		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceTLVTagsReconnectHere, config.Address(s.cfg.OSCARHost, s.cfg.ChatPort)),
						wire.NewTLV(wire.OServiceTLVTagsLoginCookie, chatLoginCookie{
							Cookie: room.Cookie,
							SessID: sess.ID(),
						}),
						wire.NewTLV(wire.OServiceTLVTagsGroupID, wire.Chat),
						wire.NewTLV(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLV(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	default:
		err := fmt.Errorf("%w. food group: %s", wire.ErrUnsupportedFoodGroup, wire.FoodGroupName(inBody.FoodGroup))
		return wire.SNACMessage{}, err
	}
}

// HostOnline initiates the BOS protocol sequence.
// It returns SNAC wire.OServiceHostOnline containing the list food groups
// supported by the BOS service.
// ChatNav is provided by BOS in addition to the standalone ChatNav service.
// AIM 4.x always creates a secondary TCP connection for ChatNav, whereas 5.x
// can use the existing BOS connection for ChatNav services.
func (s OServiceServiceForBOS) HostOnline() wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.Alert,
				wire.Buddy,
				wire.ChatNav,
				wire.Feedbag,
				wire.ICBM,
				wire.Locate,
				wire.OService,
				wire.BART,
			},
		},
	}
}

// ClientOnline runs when the current user is ready to join.
// It announces current user's arrival to users who have the current user on
// their buddy list.
func (s OServiceServiceForBOS) ClientOnline(ctx context.Context, _ wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	sess.SetSignonComplete()

	if err := broadcastArrival(ctx, sess, s.messageRelayer, s.feedbagManager, s.legacyBuddyListManager); err != nil {
		return err
	}

	// send buddy arrival events to client-side buddy list
	buddies := s.legacyBuddyListManager.Buddies(sess.ScreenName())
	for _, buddy := range buddies {
		buddySess := s.messageRelayer.RetrieveByScreenName(buddy)
		if buddySess == nil || buddySess.Invisible() {
			continue
		}
		if err := unicastArrival(ctx, buddySess, sess, s.messageRelayer, s.feedbagManager); err != nil {
			return err
		}
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

// OServiceServiceForChat provides functionality for the OService food group
// running on the Chat server.
type OServiceServiceForChat struct {
	OServiceService
	chatRegistry *state.ChatRegistry
}

// HostOnline initiates the Chat protocol sequence.
// It returns SNAC wire.OServiceHostOnline containing the list of food groups
// supported by the Chat service.
func (s OServiceServiceForChat) HostOnline() wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.OService,
				wire.Chat,
			},
		},
	}
}

// ClientOnline runs when the current user is ready to join the chat.
// Trigger the following actions:
//   - Send current user the chat room metadata
//   - Announce current user's arrival to other chat room participants
//   - Send current user the chat room participant list
func (s OServiceServiceForChat) ClientOnline(ctx context.Context, sess *state.Session) error {
	room, chatSessMgr, err := s.chatRegistry.Retrieve(sess.ChatRoomCookie())
	if err != nil {
		return err
	}
	sendChatRoomInfoUpdate(ctx, sess, chatSessMgr.(ChatMessageRelayer), room)
	alertUserJoined(ctx, sess, chatSessMgr.(ChatMessageRelayer))
	setOnlineChatUsers(ctx, sess, chatSessMgr.(ChatMessageRelayer))
	return nil
}

// NewOServiceServiceForChatNav creates a new instance of OServiceServiceForChat.
func NewOServiceServiceForChatNav(oserviceService OServiceService, chatRegistry *state.ChatRegistry) *OServiceServiceForChatNav {
	return &OServiceServiceForChatNav{
		OServiceService: oserviceService,
		chatRegistry:    chatRegistry,
	}
}

// OServiceServiceForChatNav provides functionality for the OService food group
// running on the Chat server.
type OServiceServiceForChatNav struct {
	OServiceService
	chatRegistry *state.ChatRegistry
}

// HostOnline initiates the ChatNav protocol sequence.
// It returns SNAC wire.OServiceHostOnline containing the list of food groups
// supported by the ChatNav service.
// ChatNav is provided by BOS in addition to the standalone ChatNav service.
// AIM 4.x always creates a secondary TCP connection for ChatNav, whereas 5.x
// can use the existing BOS connection for ChatNav services.
func (s OServiceServiceForChatNav) HostOnline() wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.ChatNav,
				wire.OService,
			},
		},
	}
}
