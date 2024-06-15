package foodgroup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// OServiceService provides functionality for the OService food group, which
// provides an assortment of services useful across multiple food groups.
type OServiceService struct {
	buddyUpdateBroadcaster buddyBroadcaster
	cfg                    config.Config
	logger                 *slog.Logger
	foodGroups             []uint16
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

var rateLimitSNAC = wire.SNAC_0x01_0x07_OServiceRateParamsReply{
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
			}{},
		},
	},
}

// populate the rate limit SNAC with a rule for each subgroup
func init() {
	foodGroupToSubgroup := map[uint16][]uint16{
		wire.OService: {
			wire.OServiceErr,
			wire.OServiceClientOnline,
			wire.OServiceHostOnline,
			wire.OServiceServiceRequest,
			wire.OServiceServiceResponse,
			wire.OServiceRateParamsQuery,
			wire.OServiceRateParamsReply,
			wire.OServiceRateParamsSubAdd,
			wire.OServiceRateDelParamSub,
			wire.OServiceRateParamChange,
			wire.OServicePauseReq,
			wire.OServicePauseAck,
			wire.OServiceResume,
			wire.OServiceUserInfoQuery,
			wire.OServiceUserInfoUpdate,
			wire.OServiceEvilNotification,
			wire.OServiceIdleNotification,
			wire.OServiceMigrateGroups,
			wire.OServiceMotd,
			wire.OServiceSetPrivacyFlags,
			wire.OServiceWellKnownUrls,
			wire.OServiceNoop,
			wire.OServiceClientVersions,
			wire.OServiceHostVersions,
			wire.OServiceMaxConfigQuery,
			wire.OServiceMaxConfigReply,
			wire.OServiceStoreConfig,
			wire.OServiceConfigQuery,
			wire.OServiceConfigReply,
			wire.OServiceSetUserInfoFields,
			wire.OServiceProbeReq,
			wire.OServiceProbeAck,
			wire.OServiceBartReply,
			wire.OServiceBartQuery2,
			wire.OServiceBartReply2,
		},
		wire.Locate: {
			wire.LocateErr,
			wire.LocateRightsQuery,
			wire.LocateRightsReply,
			wire.LocateSetInfo,
			wire.LocateUserInfoQuery,
			wire.LocateUserInfoReply,
			wire.LocateWatcherSubRequest,
			wire.LocateWatcherNotification,
			wire.LocateSetDirInfo,
			wire.LocateSetDirReply,
			wire.LocateGetDirInfo,
			wire.LocateGetDirReply,
			wire.LocateGroupCapabilityQuery,
			wire.LocateGroupCapabilityReply,
			wire.LocateSetKeywordInfo,
			wire.LocateSetKeywordReply,
			wire.LocateGetKeywordInfo,
			wire.LocateGetKeywordReply,
			wire.LocateFindListByEmail,
			wire.LocateFindListReply,
			wire.LocateUserInfoQuery2,
		},
		wire.Buddy: {
			wire.BuddyErr,
			wire.BuddyRightsQuery,
			wire.BuddyRightsReply,
			wire.BuddyAddBuddies,
			wire.BuddyDelBuddies,
			wire.BuddyWatcherListQuery,
			wire.BuddyWatcherListResponse,
			wire.BuddyWatcherSubRequest,
			wire.BuddyWatcherNotification,
			wire.BuddyRejectNotification,
			wire.BuddyArrived,
			wire.BuddyDeparted,
			wire.BuddyAddTempBuddies,
			wire.BuddyDelTempBuddies,
		},
		wire.ICBM: {
			wire.ICBMErr,
			wire.ICBMAddParameters,
			wire.ICBMDelParameters,
			wire.ICBMParameterQuery,
			wire.ICBMParameterReply,
			wire.ICBMChannelMsgToHost,
			wire.ICBMChannelMsgToClient,
			wire.ICBMEvilRequest,
			wire.ICBMEvilReply,
			wire.ICBMMissedCalls,
			wire.ICBMClientErr,
			wire.ICBMHostAck,
			wire.ICBMSinStored,
			wire.ICBMSinListQuery,
			wire.ICBMSinListReply,
			wire.ICBMSinRetrieve,
			wire.ICBMSinDelete,
			wire.ICBMNotifyRequest,
			wire.ICBMNotifyReply,
			wire.ICBMClientEvent,
			wire.ICBMSinReply,
		},
		wire.ChatNav: {
			wire.ChatNavErr,
			wire.ChatNavRequestChatRights,
			wire.ChatNavRequestExchangeInfo,
			wire.ChatNavRequestRoomInfo,
			wire.ChatNavRequestMoreRoomInfo,
			wire.ChatNavRequestOccupantList,
			wire.ChatNavSearchForRoom,
			wire.ChatNavCreateRoom,
			wire.ChatNavNavInfo,
		},
		wire.Chat: {
			wire.ChatErr,
			wire.ChatRoomInfoUpdate,
			wire.ChatUsersJoined,
			wire.ChatUsersLeft,
			wire.ChatChannelMsgToHost,
			wire.ChatChannelMsgToClient,
			wire.ChatEvilRequest,
			wire.ChatEvilReply,
			wire.ChatClientErr,
			wire.ChatPauseRoomReq,
			wire.ChatPauseRoomAck,
			wire.ChatResumeRoom,
			wire.ChatShowMyRow,
			wire.ChatShowRowByUsername,
			wire.ChatShowRowByNumber,
			wire.ChatShowRowByName,
			wire.ChatRowInfo,
			wire.ChatListRows,
			wire.ChatRowListInfo,
			wire.ChatMoreRows,
			wire.ChatMoveToRow,
			wire.ChatToggleChat,
			wire.ChatSendQuestion,
			wire.ChatSendComment,
			wire.ChatTallyVote,
			wire.ChatAcceptBid,
			wire.ChatSendInvite,
			wire.ChatDeclineInvite,
			wire.ChatAcceptInvite,
			wire.ChatNotifyMessage,
			wire.ChatGotoRow,
			wire.ChatStageUserJoin,
			wire.ChatStageUserLeft,
			wire.ChatUnnamedSnac22,
			wire.ChatClose,
			wire.ChatUserBan,
			wire.ChatUserUnban,
			wire.ChatJoined,
			wire.ChatUnnamedSnac27,
			wire.ChatUnnamedSnac28,
			wire.ChatUnnamedSnac29,
			wire.ChatRoomInfoOwner,
		},
		wire.BART: {
			wire.BARTErr,
			wire.BARTUploadQuery,
			wire.BARTUploadReply,
			wire.BARTDownloadQuery,
			wire.BARTDownloadReply,
			wire.BARTDownload2Query,
			wire.BARTDownload2Reply,
		},
		wire.Feedbag: {
			wire.FeedbagErr,
			wire.FeedbagRightsQuery,
			wire.FeedbagRightsReply,
			wire.FeedbagQuery,
			wire.FeedbagQueryIfModified,
			wire.FeedbagReply,
			wire.FeedbagUse,
			wire.FeedbagInsertItem,
			wire.FeedbagUpdateItem,
			wire.FeedbagDeleteItem,
			wire.FeedbagInsertClass,
			wire.FeedbagUpdateClass,
			wire.FeedbagDeleteClass,
			wire.FeedbagStatus,
			wire.FeedbagReplyNotModified,
			wire.FeedbagDeleteUser,
			wire.FeedbagStartCluster,
			wire.FeedbagEndCluster,
			wire.FeedbagAuthorizeBuddy,
			wire.FeedbagPreAuthorizeBuddy,
			wire.FeedbagPreAuthorizedBuddy,
			wire.FeedbagRemoveMe,
			wire.FeedbagRemoveMe2,
			wire.FeedbagRequestAuthorizeToHost,
			wire.FeedbagRequestAuthorizeToClient,
			wire.FeedbagRespondAuthorizeToHost,
			wire.FeedbagRespondAuthorizeToClient,
			wire.FeedbagBuddyAdded,
			wire.FeedbagRequestAuthorizeToBadog,
			wire.FeedbagRespondAuthorizeToBadog,
			wire.FeedbagBuddyAddedToBadog,
			wire.FeedbagTestSnac,
			wire.FeedbagForwardMsg,
			wire.FeedbagIsAuthRequiredQuery,
			wire.FeedbagIsAuthRequiredReply,
			wire.FeedbagRecentBuddyUpdate,
		},
		wire.BUCP: {
			wire.BUCPErr,
			wire.BUCPLoginRequest,
			wire.BUCPLoginResponse,
			wire.BUCPRegisterRequest,
			wire.BUCPChallengeRequest,
			wire.BUCPChallengeResponse,
			wire.BUCPAsasnRequest,
			wire.BUCPSecuridRequest,
			wire.BUCPRegistrationImageRequest,
		},
		wire.Alert: {
			wire.AlertErr,
			wire.AlertSetAlertRequest,
			wire.AlertSetAlertReply,
			wire.AlertGetSubsRequest,
			wire.AlertGetSubsResponse,
			wire.AlertNotifyCapabilities,
			wire.AlertNotify,
			wire.AlertGetRuleRequest,
			wire.AlertGetRuleReply,
			wire.AlertGetFeedRequest,
			wire.AlertGetFeedReply,
			wire.AlertRefreshFeed,
			wire.AlertEvent,
			wire.AlertQogSnac,
			wire.AlertRefreshFeedStock,
			wire.AlertNotifyTransport,
			wire.AlertSetAlertRequestV2,
			wire.AlertSetAlertReplyV2,
			wire.AlertTransitReply,
			wire.AlertNotifyAck,
			wire.AlertNotifyDisplayCapabilities,
			wire.AlertUserOnline,
		},
	}

	for _, foodGroup := range []uint16{
		wire.OService,
		wire.Locate,
		wire.Buddy,
		wire.ICBM,
		wire.ChatNav,
		wire.Chat,
		wire.BART,
		wire.Feedbag,
		wire.BUCP,
		wire.Alert,
	} {
		subGroups := foodGroupToSubgroup[foodGroup]
		for _, subGroup := range subGroups {
			rateLimitSNAC.RateGroups[0].Pairs = append(rateLimitSNAC.RateGroups[0].Pairs, struct {
				FoodGroup uint16
				SubGroup  uint16
			}{
				FoodGroup: foodGroup,
				SubGroup:  subGroup,
			})
		}
	}
}

// RateParamsQuery returns SNAC rate limits. It returns SNAC
// wire.OServiceRateParamsReply containing rate limits for all food groups
// supported by this server.
//
// The purpose of this method is to convey per-SNAC server-side rate limits to
// the client. The response consists of two main parts: rate classes and rate
// groups. Rate classes define limits based on specific parameters, while rate
// groups associate these limits with relevant SNAC types.
//
// The current implementation does not enforce server-side rate limiting.
// Instead, the provided values inform the client about the recommended
// client-side rate limits.
//
// The rate limit values were taken from the example SNAC dump documented here:
// https://web.archive.org/web/20221207225518/https://wiki.nina.chat/wiki/Protocols/OSCAR/SNAC/OSERVICE_RATE_PARAMS_REPLY
//
// AIM clients silently fail when they expect a rate limit rule that does not
// exist in this response. When support for a new food group is added to the
// server, update this function accordingly.
func (s OServiceService) RateParamsQuery(_ context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceRateParamsReply,
			RequestID: inFrame.RequestID,
		},
		Body: rateLimitSNAC,
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
		if status == wire.OServiceUserStatusAvailable {
			sess.SetInvisible(false)
			if err := s.buddyUpdateBroadcaster.BroadcastBuddyArrived(ctx, sess); err != nil {
				return wire.SNACMessage{}, err
			}
		}
		if status&wire.OServiceUserStatusInvisible == wire.OServiceUserStatusInvisible {
			sess.SetInvisible(true)
			if err := s.buddyUpdateBroadcaster.BroadcastBuddyDeparted(ctx, sess); err != nil {
				return wire.SNACMessage{}, err
			}
		}
		if status&wire.OServiceUserStatusDirectRequireAuth == wire.OServiceUserStatusDirectRequireAuth {
			s.logger.DebugContext(ctx, "got unsupported status", "status", status)
		}
		if status&wire.OServiceUserStatusHideIP == wire.OServiceUserStatusHideIP {
			s.logger.DebugContext(ctx, "got unsupported status", "status", status)
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
	return s.buddyUpdateBroadcaster.BroadcastBuddyArrived(ctx, sess)
}

// SetPrivacyFlags sets client privacy settings. Currently, there's no action
// to take when these flags are set. This method simply logs the flags set by
// the client.
func (s OServiceService) SetPrivacyFlags(ctx context.Context, bodyIn wire.SNAC_0x01_0x14_OServiceSetPrivacyFlags) {
	attrs := slog.Group("request",
		slog.String("food_group", wire.FoodGroupName(wire.OService)),
		slog.String("sub_group", wire.SubGroupName(wire.OService, wire.OServiceSetPrivacyFlags)))

	if bodyIn.MemberFlag() {
		s.logger.LogAttrs(ctx, slog.LevelDebug, "client set member privacy flag, but we're not going to do anything", attrs)
	}
	if bodyIn.IdleFlag() {
		s.logger.LogAttrs(ctx, slog.LevelDebug, "client set idle privacy flag, but we're not going to do anything", attrs)
	}
}

// RateParamsSubAdd exists to capture the SNAC input in unit tests to
// verify it's correctly unmarshalled.
func (s OServiceService) RateParamsSubAdd(context.Context, wire.SNAC_0x01_0x08_OServiceRateParamsSubAdd) {
}

func (s OServiceService) ServiceRequest(ctx context.Context, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x04_OServiceServiceRequest) (wire.SNACMessage, error) {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceErr,
			RequestID: frame.RequestID,
		},
		Body: wire.SNACError{
			Code: wire.ErrorCodeNotSupportedByHost,
		},
	}, nil
}

// ClientOnline informs the server that the client is ready.
func (s OServiceService) ClientOnline(ctx context.Context, bodyIn wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	s.logger.DebugContext(ctx, "client is online", "group_versions", bodyIn.GroupVersions)
	return nil
}

// HostOnline initiates the Alert protocol sequence.
// It returns SNAC wire.OServiceHostOnline containing the list of food groups
// supported by the Alert service.
// Alert is provided by BOS in addition to the standalone Alert service.
// AIM 4.x always creates a secondary TCP connection for Alert, whereas 5.x
// can use the existing BOS connection for Alert services.
func (s OServiceService) HostOnline() wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: s.foodGroups,
		},
	}
}

// NewOServiceServiceForBOS creates a new instance of OServiceServiceForBOS.
func NewOServiceServiceForBOS(
	cfg config.Config,
	messageRelayer MessageRelayer,
	legacyBuddyListManager LegacyBuddyListManager,
	logger *slog.Logger,
	cookieIssuer CookieBaker,
	cr *state.ChatRegistry,
	feedbagManager FeedbagManager,
) *OServiceServiceForBOS {
	return &OServiceServiceForBOS{
		chatRegistry:           cr,
		cookieIssuer:           cookieIssuer,
		legacyBuddyListManager: legacyBuddyListManager,
		messageRelayer:         messageRelayer,
		OServiceService: OServiceService{
			buddyUpdateBroadcaster: NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager),
			cfg:                    cfg,
			logger:                 logger,
			foodGroups: []uint16{
				wire.Alert,
				wire.Buddy,
				wire.ChatNav,
				wire.Feedbag,
				wire.ICBM,
				wire.Locate,
				wire.OService,
				wire.BART,
				wire.PermitDeny,
			},
		},
	}
}

// OServiceServiceForBOS provides functionality for the OService food group
// running on the BOS server.
type OServiceServiceForBOS struct {
	OServiceService
	chatRegistry           *state.ChatRegistry
	cookieIssuer           CookieBaker
	legacyBuddyListManager LegacyBuddyListManager
	messageRelayer         MessageRelayer
}

// chatLoginCookie represents credentials used to authenticate a user chat
// session.
type chatLoginCookie struct {
	ChatCookie string `len_prefix:"uint8"`
	ScreenName string `len_prefix:"uint8"`
}

// ServiceRequest handles service discovery, providing a host name and metadata
// for connecting to the food group service specified in inFrame.
func (s OServiceServiceForBOS) ServiceRequest(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x01_0x04_OServiceServiceRequest) (wire.SNACMessage, error) {
	switch inBody.FoodGroup {
	case wire.Alert:
		cookie, err := s.cookieIssuer.Issue([]byte(sess.IdentScreenName().String()))
		if err != nil {
			return wire.SNACMessage{}, err
		}
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.AlertPort)),
						wire.NewTLV(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLV(wire.OServiceTLVTagsGroupID, wire.Alert),
						wire.NewTLV(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLV(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	case wire.BART:
		cookie, err := s.cookieIssuer.Issue([]byte(sess.IdentScreenName().String()))
		if err != nil {
			return wire.SNACMessage{}, err
		}
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.BARTPort)),
						wire.NewTLV(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLV(wire.OServiceTLVTagsGroupID, wire.BART),
						wire.NewTLV(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLV(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	case wire.ChatNav:
		cookie, err := s.cookieIssuer.Issue([]byte(sess.IdentScreenName().String()))
		if err != nil {
			return wire.SNACMessage{}, err
		}
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.ChatNavPort)),
						wire.NewTLV(wire.OServiceTLVTagsLoginCookie, cookie),
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

		room, _, err := s.chatRegistry.Retrieve(roomSNAC.Cookie)
		if err != nil {
			return wire.SNACMessage{}, fmt.Errorf("unable to retrieve room info: %w", err)
		}

		loginCookie := chatLoginCookie{
			ChatCookie: room.Cookie,
			ScreenName: sess.IdentScreenName().String(),
		}
		buf := &bytes.Buffer{}
		if err := wire.Marshal(loginCookie, buf); err != nil {
			return wire.SNACMessage{}, err
		}
		cookie, err := s.cookieIssuer.Issue(buf.Bytes())
		if err != nil {
			return wire.SNACMessage{}, err
		}

		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.ChatPort)),
						wire.NewTLV(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLV(wire.OServiceTLVTagsGroupID, wire.Chat),
						wire.NewTLV(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLV(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	default:
		s.logger.InfoContext(ctx, "client service request for unsupported service", "food_group", wire.FoodGroupName(inBody.FoodGroup))
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: wire.ErrorCodeServiceUnavailable,
			},
		}, nil
	}
}

// ClientOnline runs when the current user is ready to join.
// It announces current user's arrival to users who have the current user on
// their buddy list.
func (s OServiceServiceForBOS) ClientOnline(ctx context.Context, _ wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	sess.SetSignonComplete()

	if err := s.buddyUpdateBroadcaster.BroadcastBuddyArrived(ctx, sess); err != nil {
		return err
	}

	// send buddy arrival events to client-side buddy list
	buddies := s.legacyBuddyListManager.Buddies(sess.IdentScreenName())
	for _, buddy := range buddies {
		buddySess := s.messageRelayer.RetrieveByScreenName(buddy)
		if buddySess == nil || buddySess.Invisible() {
			continue
		}
		if err := s.buddyUpdateBroadcaster.UnicastBuddyArrived(ctx, buddySess, sess); err != nil {
			return err
		}
	}
	return nil
}

// NewOServiceServiceForChat creates a new instance of NewOServiceServiceForChat.
func NewOServiceServiceForChat(
	cfg config.Config,
	logger *slog.Logger,
	cr *state.ChatRegistry,
	messageRelayer MessageRelayer,
	legacyBuddyListManager LegacyBuddyListManager,
	feedbagManager FeedbagManager,
) *OServiceServiceForChat {
	return &OServiceServiceForChat{
		chatRegistry: cr,
		OServiceService: OServiceService{
			buddyUpdateBroadcaster: NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager),
			cfg:                    cfg,
			logger:                 logger,
			foodGroups: []uint16{
				wire.OService,
				wire.Chat,
			},
		},
	}
}

// OServiceServiceForChat provides functionality for the OService food group
// running on the Chat server.
type OServiceServiceForChat struct {
	OServiceService
	chatRegistry *state.ChatRegistry
}

// ClientOnline runs when the current user is ready to join the chat.
// Trigger the following actions:
//   - Send current user the chat room metadata
//   - Announce current user's arrival to other chat room participants
//   - Send current user the chat room participant list
func (s OServiceServiceForChat) ClientOnline(ctx context.Context, _ wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	room, chatSessMgr, err := s.chatRegistry.Retrieve(sess.ChatRoomCookie())
	if err != nil {
		return err
	}
	sendChatRoomInfoUpdate(ctx, sess, chatSessMgr.(ChatMessageRelayer), room)
	alertUserJoined(ctx, sess, chatSessMgr.(ChatMessageRelayer))
	setOnlineChatUsers(ctx, sess, chatSessMgr.(ChatMessageRelayer))
	return nil
}

// NewOServiceServiceForChatNav creates a new instance of OServiceService for
// ChatNav.
func NewOServiceServiceForChatNav(
	cfg config.Config,
	logger *slog.Logger,
	messageRelayer MessageRelayer,
	legacyBuddyListManager LegacyBuddyListManager,
	feedbagManager FeedbagManager,
) *OServiceService {
	return &OServiceService{
		buddyUpdateBroadcaster: NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager),
		cfg:                    cfg,
		logger:                 logger,
		foodGroups: []uint16{
			wire.ChatNav,
			wire.OService,
		},
	}
}

// NewOServiceServiceForAlert creates a new instance of OServiceService for the Alert
// server.
func NewOServiceServiceForAlert(
	cfg config.Config,
	logger *slog.Logger,
	messageRelayer MessageRelayer,
	legacyBuddyListManager LegacyBuddyListManager,
	feedbagManager FeedbagManager,
) *OServiceService {
	return &OServiceService{
		buddyUpdateBroadcaster: NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager),
		cfg:                    cfg,
		logger:                 logger,
		foodGroups: []uint16{
			wire.Alert,
			wire.OService,
		},
	}
}

// NewOServiceServiceForBART creates a new instance of OServiceService for the
// BART server.
func NewOServiceServiceForBART(
	cfg config.Config,
	logger *slog.Logger,
	messageRelayer MessageRelayer,
	legacyBuddyListManager LegacyBuddyListManager,
	feedbagManager FeedbagManager,
) *OServiceService {
	return &OServiceService{
		buddyUpdateBroadcaster: NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager),
		cfg:                    cfg,
		logger:                 logger,
		foodGroups: []uint16{
			wire.BART,
			wire.OService,
		},
	}
}
