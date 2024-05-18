package wire

var foodGroupName = map[uint16]string{
	OService:    "OService",
	Locate:      "Locate",
	Buddy:       "Buddy",
	ICBM:        "ICBM",
	Advert:      "Advert",
	Invite:      "Invite",
	Admin:       "Admin",
	Popup:       "Popup",
	PermitDeny:  "PermitDeny",
	UserLookup:  "UserLookup",
	Stats:       "Stats",
	Translate:   "Translate",
	ChatNav:     "ChatNav",
	Chat:        "Chat",
	ODir:        "ODir",
	BART:        "BART",
	Feedbag:     "Feedbag",
	ICQ:         "ICQ",
	BUCP:        "BUCP",
	Alert:       "Alert",
	Plugin:      "Plugin",
	UnnamedFG24: "UnnamedFG24",
	MDir:        "MDir",
	ARS:         "ARS",
}

// FoodGroupName gets the string name of a food group. It returns "unknown" if
// the food group doesn't exist.
func FoodGroupName(foodGroup uint16) string {
	name := foodGroupName[foodGroup]
	if name == "" {
		name = "unknown"
	}
	return name
}

var subGroupName = map[uint16]map[uint16]string{
	OService: {
		OServiceErr:               "OServiceErr",
		OServiceClientOnline:      "OServiceClientOnline",
		OServiceHostOnline:        "HostOnline",
		OServiceServiceRequest:    "OServiceServiceRequest",
		OServiceServiceResponse:   "OServiceServiceResponse",
		OServiceRateParamsQuery:   "OServiceRateParamsQuery",
		OServiceRateParamsReply:   "OServiceRateParamsReply",
		OServiceRateParamsSubAdd:  "OServiceRateParamsSubAdd",
		OServiceRateDelParamSub:   "OServiceRateDelParamSub",
		OServiceRateParamChange:   "OServiceRateParamChange",
		OServicePauseReq:          "OServicePauseReq",
		OServicePauseAck:          "OServicePauseAck",
		OServiceResume:            "OServiceResume",
		OServiceUserInfoQuery:     "OServiceUserInfoQuery",
		OServiceUserInfoUpdate:    "OServiceUserInfoUpdate",
		OServiceEvilNotification:  "OServiceEvilNotification",
		OServiceIdleNotification:  "OServiceIdleNotification",
		OServiceMigrateGroups:     "OServiceMigrateGroups",
		OServiceMotd:              "OServiceMotd",
		OServiceSetPrivacyFlags:   "OServiceSetPrivacyFlags",
		OServiceWellKnownUrls:     "OServiceWellKnownUrls",
		OServiceNoop:              "OServiceNoop",
		OServiceClientVersions:    "ClientVersions",
		OServiceHostVersions:      "OServiceHostVersions",
		OServiceMaxConfigQuery:    "OServiceMaxConfigQuery",
		OServiceMaxConfigReply:    "OServiceMaxConfigReply",
		OServiceStoreConfig:       "OServiceStoreConfig",
		OServiceConfigQuery:       "OServiceConfigQuery",
		OServiceConfigReply:       "OServiceConfigReply",
		OServiceSetUserInfoFields: "OServiceSetUserInfoFields",
		OServiceProbeReq:          "OServiceProbeReq",
		OServiceProbeAck:          "OServiceProbeAck",
		OServiceBartReply:         "OServiceBartReply",
		OServiceBartQuery2:        "OServiceBartQuery2",
		OServiceBartReply2:        "OServiceBartReply2",
	},
	Locate: {
		LocateErr:                  "LocateErr",
		LocateRightsQuery:          "LocateRightsQuery",
		LocateRightsReply:          "LocateRightsReply",
		LocateSetInfo:              "LocateSetInfo",
		LocateUserInfoQuery:        "LocateUserInfoQuery",
		LocateUserInfoReply:        "LocateUserInfoReply",
		LocateWatcherSubRequest:    "LocateWatcherSubRequest",
		LocateWatcherNotification:  "LocateWatcherNotification",
		LocateSetDirInfo:           "LocateSetDirInfo",
		LocateSetDirReply:          "LocateSetDirReply",
		LocateGetDirInfo:           "LocateGetDirInfo",
		LocateGetDirReply:          "LocateGetDirReply",
		LocateGroupCapabilityQuery: "LocateGroupCapabilityQuery",
		LocateGroupCapabilityReply: "LocateGroupCapabilityReply",
		LocateSetKeywordInfo:       "LocateSetKeywordInfo",
		LocateSetKeywordReply:      "LocateSetKeywordReply",
		LocateGetKeywordInfo:       "LocateGetKeywordInfo",
		LocateGetKeywordReply:      "LocateGetKeywordReply",
		LocateFindListByEmail:      "LocateFindListByEmail",
		LocateFindListReply:        "LocateFindListReply",
		LocateUserInfoQuery2:       "LocateUserInfoQuery2",
	},
	Buddy: {
		BuddyErr:                 "BuddyErr",
		BuddyRightsQuery:         "BuddyRightsQuery",
		BuddyRightsReply:         "BuddyRightsReply",
		BuddyAddBuddies:          "BuddyAddBuddies",
		BuddyDelBuddies:          "BuddyDelBuddies",
		BuddyWatcherListQuery:    "BuddyWatcherListQuery",
		BuddyWatcherListResponse: "BuddyWatcherListResponse",
		BuddyWatcherSubRequest:   "BuddyWatcherSubRequest",
		BuddyWatcherNotification: "BuddyWatcherNotification",
		BuddyRejectNotification:  "BuddyRejectNotification",
		BuddyArrived:             "BuddyArrived",
		BuddyDeparted:            "BuddyDeparted",
		BuddyAddTempBuddies:      "BuddyAddTempBuddies",
		BuddyDelTempBuddies:      "BuddyDelTempBuddies",
	},
	ICBM: {
		ICBMErr:                "ICBMErr",
		ICBMAddParameters:      "ICBMAddParameters",
		ICBMDelParameters:      "ICBMDelParameters",
		ICBMParameterQuery:     "ICBMParameterQuery",
		ICBMParameterReply:     "ICBMParameterReply",
		ICBMChannelMsgToHost:   "ICBMChannelMsgToHost",
		ICBMChannelMsgToClient: "ICBMChannelMsgToClient",
		ICBMEvilRequest:        "ICBMEvilRequest",
		ICBMEvilReply:          "ICBMEvilReply",
		ICBMMissedCalls:        "ICBMMissedCalls",
		ICBMClientErr:          "ICBMClientErr",
		ICBMHostAck:            "ICBMHostAck",
		ICBMSinStored:          "ICBMSinStored",
		ICBMSinListQuery:       "ICBMSinListQuery",
		ICBMSinListReply:       "ICBMSinListReply",
		ICBMSinRetrieve:        "ICBMSinRetrieve",
		ICBMSinDelete:          "ICBMSinDelete",
		ICBMNotifyRequest:      "ICBMNotifyRequest",
		ICBMNotifyReply:        "ICBMNotifyReply",
		ICBMClientEvent:        "ICBMClientEvent",
		ICBMSinReply:           "ICBMSinReply",
	},
	ChatNav: {
		ChatNavErr:                 "ChatNavErr",
		ChatNavRequestChatRights:   "ChatNavRequestChatRights",
		ChatNavRequestExchangeInfo: "ChatNavRequestExchangeInfo",
		ChatNavRequestRoomInfo:     "ChatNavRequestRoomInfo",
		ChatNavRequestMoreRoomInfo: "ChatNavRequestMoreRoomInfo",
		ChatNavRequestOccupantList: "ChatNavRequestOccupantList",
		ChatNavSearchForRoom:       "ChatNavSearchForRoom",
		ChatNavCreateRoom:          "ChatNavCreateRoom",
		ChatNavNavInfo:             "ChatNavNavInfo",
	},
	Chat: {
		ChatErr:                "ChatErr",
		ChatRoomInfoUpdate:     "ChatRoomInfoUpdate",
		ChatUsersJoined:        "ChatUsersJoined",
		ChatUsersLeft:          "ChatUsersLeft",
		ChatChannelMsgToHost:   "ChatChannelMsgToHost",
		ChatChannelMsgToClient: "ChatChannelMsgToClient",
		ChatEvilRequest:        "ChatEvilRequest",
		ChatEvilReply:          "ChatEvilReply",
		ChatClientErr:          "ChatClientErr",
		ChatPauseRoomReq:       "ChatPauseRoomReq",
		ChatPauseRoomAck:       "ChatPauseRoomAck",
		ChatResumeRoom:         "ChatResumeRoom",
		ChatShowMyRow:          "ChatShowMyRow",
		ChatShowRowByUsername:  "ChatShowRowByUsername",
		ChatShowRowByNumber:    "ChatShowRowByNumber",
		ChatShowRowByName:      "ChatShowRowByName",
		ChatRowInfo:            "ChatRowInfo",
		ChatListRows:           "ChatListRows",
		ChatRowListInfo:        "ChatRowListInfo",
		ChatMoreRows:           "ChatMoreRows",
		ChatMoveToRow:          "ChatMoveToRow",
		ChatToggleChat:         "ChatToggleChat",
		ChatSendQuestion:       "ChatSendQuestion",
		ChatSendComment:        "ChatSendComment",
		ChatTallyVote:          "ChatTallyVote",
		ChatAcceptBid:          "ChatAcceptBid",
		ChatSendInvite:         "ChatSendInvite",
		ChatDeclineInvite:      "ChatDeclineInvite",
		ChatAcceptInvite:       "ChatAcceptInvite",
		ChatNotifyMessage:      "ChatNotifyMessage",
		ChatGotoRow:            "ChatGotoRow",
		ChatStageUserJoin:      "ChatStageUserJoin",
		ChatStageUserLeft:      "ChatStageUserLeft",
		ChatUnnamedSnac22:      "ChatUnnamedSnac22",
		ChatClose:              "ChatClose",
		ChatUserBan:            "ChatUserBan",
		ChatUserUnban:          "ChatUserUnban",
		ChatJoined:             "ChatJoined",
		ChatUnnamedSnac27:      "ChatUnnamedSnac27",
		ChatUnnamedSnac28:      "ChatUnnamedSnac28",
		ChatUnnamedSnac29:      "ChatUnnamedSnac29",
		ChatRoomInfoOwner:      "ChatRoomInfoOwner",
	},
	Feedbag: {
		FeedbagErr:                      "FeedbagErr",
		FeedbagRightsQuery:              "FeedbagRightsQuery",
		FeedbagRightsReply:              "FeedbagRightsReply",
		FeedbagQuery:                    "FeedbagQuery",
		FeedbagQueryIfModified:          "FeedbagQueryIfModified",
		FeedbagReply:                    "FeedbagReply",
		FeedbagUse:                      "FeedbagUse",
		FeedbagInsertItem:               "FeedbagInsertItem",
		FeedbagUpdateItem:               "FeedbagUpdateItem",
		FeedbagDeleteItem:               "FeedbagDeleteItem",
		FeedbagInsertClass:              "FeedbagInsertClass",
		FeedbagUpdateClass:              "FeedbagUpdateClass",
		FeedbagDeleteClass:              "FeedbagDeleteClass",
		FeedbagStatus:                   "FeedbagStatus",
		FeedbagReplyNotModified:         "FeedbagReplyNotModified",
		FeedbagDeleteUser:               "FeedbagDeleteUser",
		FeedbagStartCluster:             "FeedbagStartCluster",
		FeedbagEndCluster:               "FeedbagEndCluster",
		FeedbagAuthorizeBuddy:           "FeedbagAuthorizeBuddy",
		FeedbagPreAuthorizeBuddy:        "FeedbagPreAuthorizeBuddy",
		FeedbagPreAuthorizedBuddy:       "FeedbagPreAuthorizedBuddy",
		FeedbagRemoveMe:                 "FeedbagRemoveMe",
		FeedbagRemoveMe2:                "FeedbagRemoveMe2",
		FeedbagRequestAuthorizeToHost:   "FeedbagRequestAuthorizeToHost",
		FeedbagRequestAuthorizeToClient: "FeedbagRequestAuthorizeToClient",
		FeedbagRespondAuthorizeToHost:   "FeedbagRespondAuthorizeToHost",
		FeedbagRespondAuthorizeToClient: "FeedbagRespondAuthorizeToClient",
		FeedbagBuddyAdded:               "FeedbagBuddyAdded",
		FeedbagRequestAuthorizeToBadog:  "FeedbagRequestAuthorizeToBadog",
		FeedbagRespondAuthorizeToBadog:  "FeedbagRespondAuthorizeToBadog",
		FeedbagBuddyAddedToBadog:        "FeedbagBuddyAddedToBadog",
		FeedbagTestSnac:                 "FeedbagTestSnac",
		FeedbagForwardMsg:               "FeedbagForwardMsg",
		FeedbagIsAuthRequiredQuery:      "FeedbagIsAuthRequiredQuery",
		FeedbagIsAuthRequiredReply:      "FeedbagIsAuthRequiredReply",
		FeedbagRecentBuddyUpdate:        "FeedbagRecentBuddyUpdate",
	},
	Alert: {
		AlertErr:                       "AlertErr",
		AlertSetAlertRequest:           "AlertSetAlertRequest",
		AlertSetAlertReply:             "AlertSetAlertReply",
		AlertGetSubsRequest:            "AlertGetSubsRequest",
		AlertGetSubsResponse:           "AlertGetSubsResponse",
		AlertNotifyCapabilities:        "AlertNotifyCapabilities",
		AlertNotify:                    "AlertNotify",
		AlertGetRuleRequest:            "AlertGetRuleRequest",
		AlertGetRuleReply:              "AlertGetRuleReply",
		AlertGetFeedRequest:            "AlertGetFeedRequest",
		AlertGetFeedReply:              "AlertGetFeedReply",
		AlertRefreshFeed:               "AlertRefreshFeed",
		AlertEvent:                     "AlertEvent",
		AlertQogSnac:                   "AlertQogSnac",
		AlertRefreshFeedStock:          "AlertRefreshFeedStock",
		AlertNotifyTransport:           "AlertNotifyTransport",
		AlertSetAlertRequestV2:         "AlertSetAlertRequestV2",
		AlertSetAlertReplyV2:           "AlertSetAlertReplyV2",
		AlertTransitReply:              "AlertTransitReply",
		AlertNotifyAck:                 "AlertNotifyAck",
		AlertNotifyDisplayCapabilities: "AlertNotifyDisplayCapabilities",
		AlertUserOnline:                "AlertUserOnline",
	},
	BART: {
		BARTErr:            "BARTErr",
		BARTUploadQuery:    "BARTUploadQuery",
		BARTUploadReply:    "BARTUploadReply",
		BARTDownloadQuery:  "BARTDownloadQuery",
		BARTDownloadReply:  "BARTDownloadReply",
		BARTDownload2Query: "BARTDownload2Query",
		BARTDownload2Reply: "BARTDownload2Reply",
	},
	PermitDeny: {
		PermitDenyErr:                      "PermitDenyErr",
		PermitDenyRightsQuery:              "PermitDenyRightsQuery",
		PermitDenyRightsReply:              "PermitDenyRightsReply",
		PermitDenySetGroupPermitMask:       "PermitDenySetGroupPermitMask",
		PermitDenyAddPermListEntries:       "PermitDenyAddPermListEntries",
		PermitDenyDelPermListEntries:       "PermitDenyDelPermListEntries",
		PermitDenyAddDenyListEntries:       "PermitDenyAddDenyListEntries",
		PermitDenyDelDenyListEntries:       "PermitDenyDelDenyListEntries",
		PermitDenyBosErr:                   "PermitDenyBosErr",
		PermitDenyAddTempPermitListEntries: "PermitDenyAddTempPermitListEntries",
		PermitDenyDelTempPermitListEntries: "PermitDenyDelTempPermitListEntries",
	},
}

// SubGroupName gets the string name of a subgroup within a food group. It
// returns "unknown" if the subgroup doesn't exist.
func SubGroupName(foodGroup uint16, subGroup uint16) string {
	name := subGroupName[foodGroup][subGroup]
	if name == "" {
		name = "unknown"
	}
	return name
}
