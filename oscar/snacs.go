package oscar

//
// 0x01: OService
//

const (
	OServiceErr               uint16 = 0x0001
	OServiceClientOnline      uint16 = 0x0002
	OServiceHostOnline        uint16 = 0x0003
	OServiceServiceRequest    uint16 = 0x0004
	OServiceServiceResponse   uint16 = 0x0005
	OServiceRateParamsQuery   uint16 = 0x0006
	OServiceRateParamsReply   uint16 = 0x0007
	OServiceRateParamsSubAdd  uint16 = 0x0008
	OServiceRateDelParamSub   uint16 = 0x0009
	OServiceRateParamChange   uint16 = 0x000A
	OServicePauseReq          uint16 = 0x000B
	OServicePauseAck          uint16 = 0x000C
	OServiceResume            uint16 = 0x000D
	OServiceUserInfoQuery     uint16 = 0x000E
	OServiceUserInfoUpdate    uint16 = 0x000F
	OServiceEvilNotification  uint16 = 0x0010
	OServiceIdleNotification  uint16 = 0x0011
	OServiceMigrateGroups     uint16 = 0x0012
	OServiceMotd              uint16 = 0x0013
	OServiceSetPrivacyFlags   uint16 = 0x0014
	OServiceWellKnownUrls     uint16 = 0x0015
	OServiceNoop              uint16 = 0x0016
	OServiceClientVersions    uint16 = 0x0017
	OServiceHostVersions      uint16 = 0x0018
	OServiceMaxConfigQuery    uint16 = 0x0019
	OServiceMaxConfigReply    uint16 = 0x001A
	OServiceStoreConfig       uint16 = 0x001B
	OServiceConfigQuery       uint16 = 0x001C
	OServiceConfigReply       uint16 = 0x001D
	OServiceSetUserInfoFields uint16 = 0x001E
	OServiceProbeReq          uint16 = 0x001F
	OServiceProbeAck          uint16 = 0x0020
	OServiceBartReply         uint16 = 0x0021
	OServiceBartQuery2        uint16 = 0x0022
	OServiceBartReply2        uint16 = 0x0023

	OServiceTLVTagsReconnectHere uint16 = 0x05
	OServiceTLVTagsLoginCookie   uint16 = 0x06
	OServiceTLVTagsGroupID       uint16 = 0x0D
	OServiceTLVTagsSSLCertName   uint16 = 0x8D
	OServiceTLVTagsSSLState      uint16 = 0x8E
)

type SNAC_0x01_0x02_OServiceClientOnline struct {
	GroupVersions []struct {
		FoodGroup   uint16
		Version     uint16
		ToolID      uint16
		ToolVersion uint16
	}
}

type SNAC_0x01_0x03_OServiceHostOnline struct {
	FoodGroups []uint16
}

type SNAC_0x01_0x04_OServiceServiceRequest struct {
	FoodGroup uint16
	TLVRestBlock
}

type SNAC_0x01_0x04_TLVRoomInfo struct {
	Exchange       uint16
	Cookie         []byte `len_prefix:"uint8"`
	InstanceNumber uint16
}

type SNAC_0x01_0x05_OServiceServiceResponse struct {
	TLVRestBlock
}

type SNAC_0x01_0x07_OServiceRateParamsReply struct {
	RateClasses []struct {
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
	} `count_prefix:"uint16"`
	RateGroups []struct {
		ID    uint16
		Pairs []struct {
			FoodGroup uint16
			SubGroup  uint16
		} `count_prefix:"uint16"`
	}
}

type SNAC_0x01_0x08_OServiceRateParamsSubAdd struct {
	TLVRestBlock
}

type SNAC_0x01_0x0F_OServiceUserInfoUpdate struct {
	TLVUserInfo
}

type SNAC_0x01_0x10_OServiceEvilNotification struct {
	NewEvil uint16
	TLVUserInfo
}

type SNAC_0x01_0x10_OServiceEvilNotificationAnon struct {
	NewEvil uint16
}

type SNAC_0x01_0x11_OServiceIdleNotification struct {
	IdleTime uint32
}

type SNAC_0x01_0x17_OServiceClientVersions struct {
	Versions []uint16
}

type SNAC_0x01_0x18_OServiceHostVersions struct {
	Versions []uint16
}

type SNAC_0x01_0x1E_OServiceSetUserInfoFields struct {
	TLVRestBlock
}

//
// 0x02: Locate
//

const (
	LocateErr                  uint16 = 0x0001
	LocateRightsQuery          uint16 = 0x0002
	LocateRightsReply          uint16 = 0x0003
	LocateSetInfo              uint16 = 0x0004
	LocateUserInfoQuery        uint16 = 0x0005
	LocateUserInfoReply        uint16 = 0x0006
	LocateWatcherSubRequest    uint16 = 0x0007
	LocateWatcherNotification  uint16 = 0x0008
	LocateSetDirInfo           uint16 = 0x0009
	LocateSetDirReply          uint16 = 0x000A
	LocateGetDirInfo           uint16 = 0x000B
	LocateGetDirReply          uint16 = 0x000C
	LocateGroupCapabilityQuery uint16 = 0x000D
	LocateGroupCapabilityReply uint16 = 0x000E
	LocateSetKeywordInfo       uint16 = 0x000F
	LocateSetKeywordReply      uint16 = 0x0010
	LocateGetKeywordInfo       uint16 = 0x0011
	LocateGetKeywordReply      uint16 = 0x0012
	LocateFindListByEmail      uint16 = 0x0013
	LocateFindListReply        uint16 = 0x0014
	LocateUserInfoQuery2       uint16 = 0x0015

	LocateType2Sig          uint32 = 0x00000001
	LocateType2Unavailable  uint32 = 0x00000002
	LocateType2Capabilities uint32 = 0x00000004
	LocateType2Certs        uint32 = 0x00000008
	LocateType2HtmlInfo     uint32 = 0x00000400

	LocateTLVTagsInfoSigMime         uint16 = 0x01
	LocateTLVTagsInfoSigData         uint16 = 0x02
	LocateTLVTagsInfoUnavailableMime uint16 = 0x03
	LocateTLVTagsInfoUnavailableData uint16 = 0x04
	LocateTLVTagsInfoCapabilities    uint16 = 0x05
	LocateTLVTagsInfoCerts           uint16 = 0x06
	LocateTLVTagsInfoSigTime         uint16 = 0x0A
	LocateTLVTagsInfoUnavailableTime uint16 = 0x0B
	LocateTLVTagsInfoSupportHostSig  uint16 = 0x0C
	LocateTLVTagsInfoHtmlInfoData    uint16 = 0x0E
	LocateTLVTagsInfoHtmlInfoType    uint16 = 0x0D
)

type SNAC_0x02_0x03_LocateRightsReply struct {
	TLVRestBlock
}

type SNAC_0x02_0x04_LocateSetInfo struct {
	TLVRestBlock
}

type SNAC_0x02_0x06_LocateUserInfoReply struct {
	TLVUserInfo
	LocateInfo TLVRestBlock
}

type SNAC_0x02_0x09_LocateSetDirInfo struct {
	TLVRestBlock
}

type SNAC_0x02_0x0A_LocateSetDirReply struct {
	Result uint16
}

type SNAC_0x02_0x0B_LocateGetDirInfo struct {
	WatcherScreenNames string `len_prefix:"uint8"`
}

type SNAC_0x02_0x0F_LocateSetKeywordInfo struct {
	TLVRestBlock
}

type SNAC_0x02_0x10_LocateSetKeywordReply struct {
	Unknown uint16
}

type SNAC_0x02_0x15_LocateUserInfoQuery2 struct {
	Type2      uint32
	ScreenName string `len_prefix:"uint8"`
}

func (s SNAC_0x02_0x15_LocateUserInfoQuery2) RequestProfile() bool {
	return s.Type2&LocateType2Sig == LocateType2Sig
}

func (s SNAC_0x02_0x15_LocateUserInfoQuery2) RequestAwayMessage() bool {
	return s.Type2&LocateType2Unavailable == LocateType2Unavailable
}

//
// 0x03: Buddy
//

type SNAC_0x03_0x02_BuddyRightsQuery struct {
	TLVRestBlock
}

type SNAC_0x03_0x03_BuddyRightsReply struct {
	TLVRestBlock
}

type SNAC_0x03_0x0A_BuddyArrived struct {
	TLVUserInfo
}

type SNAC_0x03_0x0B_BuddyDeparted struct {
	TLVUserInfo
}

//
// 0x04: ICBM
//

const (
	ICBMErr                uint16 = 0x0001
	ICBMAddParameters      uint16 = 0x0002
	ICBMDelParameters      uint16 = 0x0003
	ICBMParameterQuery     uint16 = 0x0004
	ICBMParameterReply     uint16 = 0x0005
	ICBMChannelMsgToHost   uint16 = 0x0006
	ICBMChannelMsgToclient uint16 = 0x0007
	ICBMEvilRequest        uint16 = 0x0008
	ICBMEvilReply          uint16 = 0x0009
	ICBMMissedCalls        uint16 = 0x000A
	ICBMClientErr          uint16 = 0x000B
	ICBMHostAck            uint16 = 0x000C
	ICBMSinStored          uint16 = 0x000D
	ICBMSinListQuery       uint16 = 0x000E
	ICBMSinListReply       uint16 = 0x000F
	ICBMSinRetrieve        uint16 = 0x0010
	ICBMSinDelete          uint16 = 0x0011
	ICBMNotifyRequest      uint16 = 0x0012
	ICBMNotifyReply        uint16 = 0x0013
	ICBMClientEvent        uint16 = 0x0014
	ICBMSinReply           uint16 = 0x0017

	ICBMTLVTagRequestHostAck uint16 = 0x03
	ICBMTLVTagsWantEvents    uint16 = 0x0B
)

type SNAC_0x04_0x02_ICBMAddParameters struct {
	Channel              uint16
	ICBMFlags            uint32
	MaxIncomingICBMLen   uint16
	MaxSourceEvil        uint16
	MaxDestinationEvil   uint16
	MinInterICBMInterval uint32
}

type SNAC_0x04_0x05_ICBMParameterReply struct {
	MaxSlots             uint16
	ICBMFlags            uint32
	MaxIncomingICBMLen   uint16
	MaxSourceEvil        uint16
	MaxDestinationEvil   uint16
	MinInterICBMInterval uint32
}

type SNAC_0x04_0x06_ICBMChannelMsgToHost struct {
	Cookie     [8]byte
	ChannelID  uint16
	ScreenName string `len_prefix:"uint8"`
	TLVRestBlock
}

type SNAC_0x04_0x07_ICBMChannelMsgToClient struct {
	Cookie    [8]byte
	ChannelID uint16
	TLVUserInfo
	TLVRestBlock
}

type SNAC_0x04_0x08_ICBMEvilRequest struct {
	SendAs     uint16
	ScreenName string `len_prefix:"uint8"`
}

type SNAC_0x04_0x09_ICBMEvilReply struct {
	EvilDeltaApplied uint16
	UpdatedEvilValue uint16
}

type SNAC_0x04_0x0B_ICBMClientErr struct {
	Cookie     [8]byte
	ChannelID  uint16
	ScreenName string `len_prefix:"uint8"`
	Code       uint16
	ErrInfo    []byte
}

type SNAC_0x04_0x0C_ICBMHostAck struct {
	Cookie     [8]byte
	ChannelID  uint16
	ScreenName string `len_prefix:"uint8"`
}

type SNAC_0x04_0x14_ICBMClientEvent struct {
	Cookie     [8]byte
	ChannelID  uint16
	ScreenName string `len_prefix:"uint8"`
	Event      uint16
}

//
// 0x09: PD
//

type SNAC_0x09_0x03_PDRightsReply struct {
	TLVRestBlock
}

//
// 0x0D: ChatNav
//

const (
	ChatNavTLVRedirect     uint16 = 0x01
	ChatNavTLVMaxRooms     uint16 = 0x02
	ChatNavTLVExchangeInfo uint16 = 0x03
	ChatNavTLVRoomInfo     uint16 = 0x04
)

type SNAC_0x0D_0x04_ChatNavRequestRoomInfo struct {
	Exchange       uint16
	Cookie         []byte `len_prefix:"uint8"`
	InstanceNumber uint16
	DetailLevel    uint8
}

type SNAC_0x0D_0x09_ChatNavNavInfo struct {
	TLVRestBlock
}

type SNAC_0x0D_0x09_TLVExchangeInfo struct {
	Identifier uint16
	TLVBlock
}

//
// 0x0E: Chat
//

const (
	ChatTLVPublicWhisperFlag    uint16 = 0x01
	ChatTLVSenderInformation    uint16 = 0x03
	ChatTLVEnableReflectionFlag uint16 = 0x06
	ChatTLVRoomName             uint16 = 0xD3
)

type SNAC_0x0E_0x02_ChatRoomInfoUpdate struct {
	Exchange       uint16
	Cookie         string `len_prefix:"uint8"`
	InstanceNumber uint16
	DetailLevel    uint8
	TLVBlock
}

type SNAC_0x0E_0x03_ChatUsersJoined struct {
	Users []TLVUserInfo
}

type SNAC_0x0E_0x04_ChatUsersLeft struct {
	Users []TLVUserInfo
}

type SNAC_0x0E_0x05_ChatChannelMsgToHost struct {
	Cookie  uint64
	Channel uint16
	TLVRestBlock
}

type SNAC_0x0E_0x06_ChatChannelMsgToClient struct {
	Cookie  uint64
	Channel uint16
	TLVRestBlock
}

//
// 0x13: Feedbag
//

const (
	FeedbagClassIdBuddy            uint16 = 0x0000
	FeedbagClassIdGroup            uint16 = 0x0001
	FeedbagClassIDPermit           uint16 = 0x0002
	FeedbagClassIDDeny             uint16 = 0x0003
	FeedbagClassIdPdinfo           uint16 = 0x0004
	FeedbagClassIdBuddyPrefs       uint16 = 0x0005
	FeedbagClassIdNonbuddy         uint16 = 0x0006
	FeedbagClassIdTpaProvider      uint16 = 0x0007
	FeedbagClassIdTpaSubscription  uint16 = 0x0008
	FeedbagClassIdClientPrefs      uint16 = 0x0009
	FeedbagClassIdStock            uint16 = 0x000A
	FeedbagClassIdWeather          uint16 = 0x000B
	FeedbagClassIdWatchList        uint16 = 0x000D
	FeedbagClassIdIgnoreList       uint16 = 0x000E
	FeedbagClassIdDateTime         uint16 = 0x000F
	FeedbagClassIdExternalUser     uint16 = 0x0010
	FeedbagClassIdRootCreator      uint16 = 0x0011
	FeedbagClassIdFish             uint16 = 0x0012
	FeedbagClassIdImportTimestamp  uint16 = 0x0013
	FeedbagClassIdBart             uint16 = 0x0014
	FeedbagClassIdRbOrder          uint16 = 0x0015
	FeedbagClassIdPersonality      uint16 = 0x0016
	FeedbagClassIdAlProf           uint16 = 0x0017
	FeedbagClassIdAlInfo           uint16 = 0x0018
	FeedbagClassIdInteraction      uint16 = 0x0019
	FeedbagClassIdVanityInfo       uint16 = 0x001D
	FeedbagClassIdFavoriteLocation uint16 = 0x001E
	FeedbagClassIdBartPdinfo       uint16 = 0x001F
	FeedbagClassIdCustomEmoticons  uint16 = 0x0024
	FeedbagClassIdMaxPredefined    uint16 = 0x0024
	FeedbagClassIdXIcqStatusNote   uint16 = 0x015C
	FeedbagClassIdMin              uint16 = 0x0400

	FeedbagAttributesShared                  uint16 = 0x0064
	FeedbagAttributesInvited                 uint16 = 0x0065
	FeedbagAttributesPending                 uint16 = 0x0066
	FeedbagAttributesTimeT                   uint16 = 0x0067
	FeedbagAttributesDenied                  uint16 = 0x0068
	FeedbagAttributesSwimIndex               uint16 = 0x0069
	FeedbagAttributesRecentBuddy             uint16 = 0x006A
	FeedbagAttributesAutoBot                 uint16 = 0x006B
	FeedbagAttributesInteraction             uint16 = 0x006D
	FeedbagAttributesMegaBot                 uint16 = 0x006F
	FeedbagAttributesOrder                   uint16 = 0x00C8
	FeedbagAttributesBuddyPrefs              uint16 = 0x00C9
	FeedbagAttributesPdMode                  uint16 = 0x00CA
	FeedbagAttributesPdMask                  uint16 = 0x00CB
	FeedbagAttributesPdFlags                 uint16 = 0x00CC
	FeedbagAttributesClientPrefs             uint16 = 0x00CD
	FeedbagAttributesLanguage                uint16 = 0x00CE
	FeedbagAttributesFishUri                 uint16 = 0x00CF
	FeedbagAttributesWirelessPdMode          uint16 = 0x00D0
	FeedbagAttributesWirelessIgnoreMode      uint16 = 0x00D1
	FeedbagAttributesFishPdMode              uint16 = 0x00D2
	FeedbagAttributesFishIgnoreMode          uint16 = 0x00D3
	FeedbagAttributesCreateTime              uint16 = 0x00D4
	FeedbagAttributesBartInfo                uint16 = 0x00D5
	FeedbagAttributesBuddyPrefsValid         uint16 = 0x00D6
	FeedbagAttributesBuddyPrefs2             uint16 = 0x00D7
	FeedbagAttributesBuddyPrefs2Valid        uint16 = 0x00D8
	FeedbagAttributesBartList                uint16 = 0x00D9
	FeedbagAttributesArriveSound             uint16 = 0x012C
	FeedbagAttributesLeaveSound              uint16 = 0x012D
	FeedbagAttributesImage                   uint16 = 0x012E
	FeedbagAttributesColorBg                 uint16 = 0x012F
	FeedbagAttributesColorFg                 uint16 = 0x0130
	FeedbagAttributesAlias                   uint16 = 0x0131
	FeedbagAttributesPassword                uint16 = 0x0132
	FeedbagAttributesDisabled                uint16 = 0x0133
	FeedbagAttributesCollapsed               uint16 = 0x0134
	FeedbagAttributesUrl                     uint16 = 0x0135
	FeedbagAttributesActiveList              uint16 = 0x0136
	FeedbagAttributesEmailAddr               uint16 = 0x0137
	FeedbagAttributesPhoneNumber             uint16 = 0x0138
	FeedbagAttributesCellPhoneNumber         uint16 = 0x0139
	FeedbagAttributesSmsPhoneNumber          uint16 = 0x013A
	FeedbagAttributesWireless                uint16 = 0x013B
	FeedbagAttributesNote                    uint16 = 0x013C
	FeedbagAttributesAlertPrefs              uint16 = 0x013D
	FeedbagAttributesBudalertSound           uint16 = 0x013E
	FeedbagAttributesStockalertValue         uint16 = 0x013F
	FeedbagAttributesTpalertEditUrl          uint16 = 0x0140
	FeedbagAttributesTpalertDeleteUrl        uint16 = 0x0141
	FeedbagAttributesTpprovMorealertsUrl     uint16 = 0x0142
	FeedbagAttributesFish                    uint16 = 0x0143
	FeedbagAttributesXunconfirmedxLastAccess uint16 = 0x0145
	FeedbagAttributesImSent                  uint16 = 0x0150
	FeedbagAttributesOnlineTime              uint16 = 0x0151
	FeedbagAttributesAwayMsg                 uint16 = 0x0152
	FeedbagAttributesImReceived              uint16 = 0x0153
	FeedbagAttributesBuddyfeedView           uint16 = 0x0154
	FeedbagAttributesWorkPhoneNumber         uint16 = 0x0158
	FeedbagAttributesOtherPhoneNumber        uint16 = 0x0159
	FeedbagAttributesWebPdMode               uint16 = 0x015F
	FeedbagAttributesFirstCreationTimeXc     uint16 = 0x0167
	FeedbagAttributesPdModeXc                uint16 = 0x016E

	FeedbagErr                      uint16 = 0x0001
	FeedbagRightsQuery              uint16 = 0x0002
	FeedbagRightsReply              uint16 = 0x0003
	FeedbagQuery                    uint16 = 0x0004
	FeedbagQueryIfModified          uint16 = 0x0005
	FeedbagReply                    uint16 = 0x0006
	FeedbagUse                      uint16 = 0x0007
	FeedbagInsertItem               uint16 = 0x0008
	FeedbagUpdateItem               uint16 = 0x0009
	FeedbagDeleteItem               uint16 = 0x000A
	FeedbagInsertClass              uint16 = 0x000B
	FeedbagUpdateClass              uint16 = 0x000C
	FeedbagDeleteClass              uint16 = 0x000D
	FeedbagStatus                   uint16 = 0x000E
	FeedbagReplyNotModified         uint16 = 0x000F
	FeedbagDeleteUser               uint16 = 0x0010
	FeedbagStartCluster             uint16 = 0x0011
	FeedbagEndCluster               uint16 = 0x0012
	FeedbagAuthorizeBuddy           uint16 = 0x0013
	FeedbagPreAuthorizeBuddy        uint16 = 0x0014
	FeedbagPreAuthorizedBuddy       uint16 = 0x0015
	FeedbagRemoveMe                 uint16 = 0x0016
	FeedbagRemoveMe2                uint16 = 0x0017
	FeedbagRequestAuthorizeToHost   uint16 = 0x0018
	FeedbagRequestAuthorizeToClient uint16 = 0x0019
	FeedbagRespondAuthorizeToHost   uint16 = 0x001A
	FeedbagRespondAuthorizeToClient uint16 = 0x001B
	FeedbagBuddyAdded               uint16 = 0x001C
	FeedbagRequestAuthorizeToBadog  uint16 = 0x001D
	FeedbagRespondAuthorizeToBadog  uint16 = 0x001E
	FeedbagBuddyAddedToBadog        uint16 = 0x001F
	FeedbagTestSnac                 uint16 = 0x0021
	FeedbagForwardMsg               uint16 = 0x0022
	FeedbagIsAuthRequiredQuery      uint16 = 0x0023
	FeedbagIsAuthRequiredReply      uint16 = 0x0024
	FeedbagRecentBuddyUpdate        uint16 = 0x0025
)

type SNAC_0x13_0x02_FeedbagRightsQuery struct {
	TLVRestBlock
}

type SNAC_0x13_0x03_FeedbagRightsReply struct {
	TLVRestBlock
}

type SNAC_0x13_0x05_FeedbagQueryIfModified struct {
	LastUpdate uint32
	Count      uint8
}

type SNAC_0x13_0x06_FeedbagReply struct {
	Version    uint8
	Items      []FeedbagItem `count_prefix:"uint16"`
	LastUpdate uint32
}

type SNAC_0x13_0x08_FeedbagInsertItem struct {
	Items []FeedbagItem
}

type SNAC_0x13_0x09_FeedbagUpdateItem struct {
	Items []FeedbagItem
}

type SNAC_0x13_0x0A_FeedbagDeleteItem struct {
	Items []FeedbagItem
}

type SNAC_0x13_0x0E_FeedbagStatus struct {
	Results []uint16
}

type SNAC_0x13_0x11_FeedbagStartCluster struct {
	TLVRestBlock
}

//
// 0x17: BUCP
//

type SNAC_0x17_0x02_BUCPLoginRequest struct {
	TLVRestBlock
}

type SNAC_0x17_0x03_BUCPLoginResponse struct {
	TLVRestBlock
}

type SNAC_0x17_0x06_BUCPChallengeRequest struct {
	TLVRestBlock
}

type SNAC_0x17_0x07_BUCPChallengeResponse struct {
	AuthKey string `len_prefix:"uint16"`
}

type SnacOServiceErr struct {
	Code uint16
}

type TLVUserInfo struct {
	ScreenName   string `len_prefix:"uint8"`
	WarningLevel uint16
	TLVBlock
}

type FeedbagItem struct {
	Name    string `len_prefix:"uint16"`
	GroupID uint16
	ItemID  uint16
	ClassID uint16
	TLVLBlock
}
