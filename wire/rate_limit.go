package wire

import "time"

type RateClass struct {
	ID              RateLimitClassID
	WindowSize      int64
	ClearLevel      int64
	AlertLevel      int64
	LimitLevel      int64
	DisconnectLevel int64
	MaxLevel        int64
}

type (
	RateLimitStatus  int
	RateLimitClassID uint16
)

const (
	RateLimitStatusLimited    RateLimitStatus = 1 // You're currently rate-limited
	RateLimitStatusAlert      RateLimitStatus = 2 // You're close to being rate-limited
	RateLimitStatusClear      RateLimitStatus = 3 // You're under the rate limit; all good
	RateLimitStatusDisconnect RateLimitStatus = 4 // You're under the rate limit; all good
)

var RateGroups = map[uint16]map[uint16]RateLimitClassID{
	OService: {
		OServiceErr:               1,
		OServiceClientOnline:      1,
		OServiceHostOnline:        1,
		OServiceServiceRequest:    1,
		OServiceServiceResponse:   1,
		OServiceRateParamsQuery:   1,
		OServiceRateParamsReply:   1,
		OServiceRateParamsSubAdd:  1,
		OServiceRateDelParamSub:   1,
		OServiceRateParamChange:   1,
		OServicePauseReq:          1,
		OServicePauseAck:          1,
		OServiceResume:            1,
		OServiceUserInfoQuery:     1,
		OServiceUserInfoUpdate:    1,
		OServiceEvilNotification:  1,
		OServiceIdleNotification:  1,
		OServiceMigrateGroups:     1,
		OServiceMotd:              1,
		OServiceSetPrivacyFlags:   1,
		OServiceWellKnownUrls:     1,
		OServiceNoop:              1,
		OServiceClientVersions:    1,
		OServiceHostVersions:      1,
		OServiceMaxConfigQuery:    1,
		OServiceMaxConfigReply:    1,
		OServiceStoreConfig:       1,
		OServiceConfigQuery:       1,
		OServiceConfigReply:       1,
		OServiceSetUserInfoFields: 1,
		OServiceProbeReq:          1,
		OServiceProbeAck:          1,
		OServiceBartReply:         1,
		OServiceBartQuery2:        1,
		OServiceBartReply2:        1,
	},
	Locate: {
		LocateErr:                  1,
		LocateRightsQuery:          1,
		LocateRightsReply:          1,
		LocateSetInfo:              1,
		LocateUserInfoQuery:        1,
		LocateUserInfoReply:        1,
		LocateWatcherSubRequest:    1,
		LocateWatcherNotification:  1,
		LocateSetDirInfo:           1,
		LocateSetDirReply:          1,
		LocateGetDirInfo:           1,
		LocateGetDirReply:          1,
		LocateGroupCapabilityQuery: 1,
		LocateGroupCapabilityReply: 1,
		LocateSetKeywordInfo:       1,
		LocateSetKeywordReply:      1,
		LocateGetKeywordInfo:       1,
		LocateGetKeywordReply:      1,
		LocateFindListByEmail:      1,
		LocateFindListReply:        1,
		LocateUserInfoQuery2:       1,
	},
	Buddy: {
		BuddyErr:                 1,
		BuddyRightsQuery:         1,
		BuddyRightsReply:         1,
		BuddyAddBuddies:          1,
		BuddyDelBuddies:          1,
		BuddyWatcherListQuery:    1,
		BuddyWatcherListResponse: 1,
		BuddyWatcherSubRequest:   1,
		BuddyWatcherNotification: 1,
		BuddyRejectNotification:  1,
		BuddyArrived:             1,
		BuddyDeparted:            1,
		BuddyAddTempBuddies:      1,
		BuddyDelTempBuddies:      1,
	},
	ICBM: {
		ICBMErr:                1,
		ICBMAddParameters:      1,
		ICBMDelParameters:      1,
		ICBMParameterQuery:     1,
		ICBMParameterReply:     1,
		ICBMChannelMsgToHost:   3,
		ICBMChannelMsgToClient: 1,
		ICBMEvilRequest:        1,
		ICBMEvilReply:          1,
		ICBMMissedCalls:        1,
		ICBMClientErr:          1,
		ICBMHostAck:            1,
		ICBMSinStored:          1,
		ICBMSinListQuery:       1,
		ICBMSinListReply:       1,
		ICBMSinRetrieve:        1,
		ICBMSinDelete:          1,
		ICBMNotifyRequest:      1,
		ICBMNotifyReply:        1,
		ICBMClientEvent:        1,
		ICBMSinReply:           1,
	},
	Invite: {
		InviteRequestQuery: 1,
	},
	ChatNav: {
		ChatNavErr:                 1,
		ChatNavRequestChatRights:   1,
		ChatNavRequestExchangeInfo: 1,
		ChatNavRequestRoomInfo:     1,
		ChatNavRequestMoreRoomInfo: 1,
		ChatNavRequestOccupantList: 1,
		ChatNavSearchForRoom:       1,
		ChatNavCreateRoom:          1,
		ChatNavNavInfo:             1,
	},
	Chat: {
		ChatErr:                1,
		ChatRoomInfoUpdate:     1,
		ChatUsersJoined:        1,
		ChatUsersLeft:          1,
		ChatChannelMsgToHost:   1,
		ChatChannelMsgToClient: 1,
		ChatEvilRequest:        1,
		ChatEvilReply:          1,
		ChatClientErr:          1,
		ChatPauseRoomReq:       1,
		ChatPauseRoomAck:       1,
		ChatResumeRoom:         1,
		ChatShowMyRow:          1,
		ChatShowRowByUsername:  1,
		ChatShowRowByNumber:    1,
		ChatShowRowByName:      1,
		ChatRowInfo:            1,
		ChatListRows:           1,
		ChatRowListInfo:        1,
		ChatMoreRows:           1,
		ChatMoveToRow:          1,
		ChatToggleChat:         1,
		ChatSendQuestion:       1,
		ChatSendComment:        1,
		ChatTallyVote:          1,
		ChatAcceptBid:          1,
		ChatSendInvite:         1,
		ChatDeclineInvite:      1,
		ChatAcceptInvite:       1,
		ChatNotifyMessage:      1,
		ChatGotoRow:            1,
		ChatStageUserJoin:      1,
		ChatStageUserLeft:      1,
		ChatUnnamedSnac22:      1,
		ChatClose:              1,
		ChatUserBan:            1,
		ChatUserUnban:          1,
		ChatJoined:             1,
		ChatUnnamedSnac27:      1,
		ChatUnnamedSnac28:      1,
		ChatUnnamedSnac29:      1,
		ChatRoomInfoOwner:      1,
	},
	BART: {
		BARTErr:            1,
		BARTUploadQuery:    1,
		BARTUploadReply:    1,
		BARTDownloadQuery:  1,
		BARTDownloadReply:  1,
		BARTDownload2Query: 1,
		BARTDownload2Reply: 1,
	},
	Feedbag: {
		FeedbagErr:                      1,
		FeedbagRightsQuery:              1,
		FeedbagRightsReply:              1,
		FeedbagQuery:                    1,
		FeedbagQueryIfModified:          1,
		FeedbagReply:                    1,
		FeedbagUse:                      1,
		FeedbagInsertItem:               1,
		FeedbagUpdateItem:               1,
		FeedbagDeleteItem:               1,
		FeedbagInsertClass:              1,
		FeedbagUpdateClass:              1,
		FeedbagDeleteClass:              1,
		FeedbagStatus:                   1,
		FeedbagReplyNotModified:         1,
		FeedbagDeleteUser:               1,
		FeedbagStartCluster:             1,
		FeedbagEndCluster:               1,
		FeedbagAuthorizeBuddy:           1,
		FeedbagPreAuthorizeBuddy:        1,
		FeedbagPreAuthorizedBuddy:       1,
		FeedbagRemoveMe:                 1,
		FeedbagRemoveMe2:                1,
		FeedbagRequestAuthorizeToHost:   1,
		FeedbagRequestAuthorizeToClient: 1,
		FeedbagRespondAuthorizeToHost:   1,
		FeedbagRespondAuthorizeToClient: 1,
		FeedbagBuddyAdded:               1,
		FeedbagRequestAuthorizeToBadog:  1,
		FeedbagRespondAuthorizeToBadog:  1,
		FeedbagBuddyAddedToBadog:        1,
		FeedbagTestSnac:                 1,
		FeedbagForwardMsg:               1,
		FeedbagIsAuthRequiredQuery:      1,
		FeedbagIsAuthRequiredReply:      1,
		FeedbagRecentBuddyUpdate:        1,
	},
	BUCP: {
		BUCPErr:                      1,
		BUCPLoginRequest:             1,
		BUCPLoginResponse:            1,
		BUCPRegisterRequest:          1,
		BUCPChallengeRequest:         1,
		BUCPChallengeResponse:        1,
		BUCPAsasnRequest:             1,
		BUCPSecuridRequest:           1,
		BUCPRegistrationImageRequest: 1,
	},
	Alert: {
		AlertErr:                       1,
		AlertSetAlertRequest:           1,
		AlertSetAlertReply:             1,
		AlertGetSubsRequest:            1,
		AlertGetSubsResponse:           1,
		AlertNotifyCapabilities:        1,
		AlertNotify:                    1,
		AlertGetRuleRequest:            1,
		AlertGetRuleReply:              1,
		AlertGetFeedRequest:            1,
		AlertGetFeedReply:              1,
		AlertRefreshFeed:               1,
		AlertEvent:                     1,
		AlertQogSnac:                   1,
		AlertRefreshFeedStock:          1,
		AlertNotifyTransport:           1,
		AlertSetAlertRequestV2:         1,
		AlertSetAlertReplyV2:           1,
		AlertTransitReply:              1,
		AlertNotifyAck:                 1,
		AlertNotifyDisplayCapabilities: 1,
		AlertUserOnline:                1,
	},
	ICQ: {
		ICQErr:     1,
		ICQDBQuery: 1,
		ICQDBReply: 1,
	},
	PermitDeny: {
		PermitDenyErr:                      1,
		PermitDenyRightsQuery:              1,
		PermitDenyRightsReply:              1,
		PermitDenySetGroupPermitMask:       1,
		PermitDenyAddPermListEntries:       1,
		PermitDenyDelPermListEntries:       1,
		PermitDenyAddDenyListEntries:       1,
		PermitDenyDelDenyListEntries:       1,
		PermitDenyBosErr:                   1,
		PermitDenyAddTempPermitListEntries: 1,
		PermitDenyDelTempPermitListEntries: 1,
	},
	ODir: {
		ODirErr:              1,
		ODirInfoQuery:        1,
		ODirInfoReply:        1,
		ODirKeywordListQuery: 1,
		ODirKeywordListReply: 1,
	},
	UserLookup: {
		UserLookupFindByEmail: 1,
	},
}

var RateClasses = []RateClass{
	{
		ID:              1,
		WindowSize:      80,
		ClearLevel:      2500,
		AlertLevel:      2000,
		LimitLevel:      1500,
		DisconnectLevel: 800,
		MaxLevel:        6000,
	},
	{
		ID:              2,
		WindowSize:      80,
		ClearLevel:      3000,
		AlertLevel:      2000,
		LimitLevel:      1500,
		DisconnectLevel: 3000,
		MaxLevel:        6000,
	},
	{
		ID:              3,
		WindowSize:      20,
		ClearLevel:      5100,
		AlertLevel:      5000,
		LimitLevel:      4000,
		DisconnectLevel: 3000,
		MaxLevel:        6000,
	},
	{
		ID:              4,
		WindowSize:      20,
		ClearLevel:      5500,
		AlertLevel:      5300,
		LimitLevel:      4200,
		DisconnectLevel: 3000,
		MaxLevel:        8000,
	},
	{
		ID:              5,
		WindowSize:      10,
		ClearLevel:      5500,
		AlertLevel:      5300,
		LimitLevel:      4200,
		DisconnectLevel: 3000,
		MaxLevel:        8000,
	},
}

func RateClassLookup(foodGroup uint16, subGroup uint16) (RateClass, bool) {
	group, ok := RateGroups[foodGroup]
	if !ok {
		return RateClass{}, false
	}
	class, ok := group[subGroup]
	if !ok {
		return RateClass{}, false
	}
	return RateClasses[class-1], true
}

// CheckRateLimit calculates moving average
func CheckRateLimit(last time.Time, now time.Time, class RateClass, curAvg int64) (status RateLimitStatus, newAvg int64) {
	delta := now.Sub(last).Milliseconds()

	//curAvg = (curAvg * (class.WindowSize - 1) / class.WindowSize) + (delta / class.WindowSize)
	curAvg = (curAvg*(class.WindowSize-1) + delta) / class.WindowSize

	if curAvg > class.MaxLevel {
		curAvg = class.MaxLevel
	}

	switch {
	case curAvg < class.DisconnectLevel:
		return RateLimitStatusDisconnect, curAvg
	case curAvg < class.LimitLevel:
		return RateLimitStatusLimited, curAvg
	case curAvg < class.AlertLevel:
		return RateLimitStatusAlert, curAvg
	}

	return RateLimitStatusClear, curAvg
}
