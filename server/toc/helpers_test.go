package toc

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type addBuddiesParams []struct {
	me     state.IdentScreenName
	inBody wire.SNAC_0x03_0x04_BuddyAddBuddies
	err    error
}

type broadcastBuddyDepartedParams []struct {
	me  state.IdentScreenName
	err error
}

type delBuddiesParams []struct {
	me     state.IdentScreenName
	inBody wire.SNAC_0x03_0x05_BuddyDelBuddies
	err    error
}

type rightsQueryParams []struct {
	inFrame wire.SNACFrame
	msg     wire.SNACMessage
}

type buddyParams struct {
	addBuddiesParams
	broadcastBuddyDepartedParams
	delBuddiesParams
	rightsQueryParams
}

type chatParams struct {
	channelMsgToHostParamsChat
}

type createRoomParams []struct {
	me     state.IdentScreenName
	inBody wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate
	msg    wire.SNACMessage
	err    error
}

type exchangeInfoParams []struct {
	inFrame wire.SNACFrame
	inBody  wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo
	msg     wire.SNACMessage
	err     error
}

type requestChatRightsParams []struct {
	inFrame wire.SNACFrame
	msg     wire.SNACMessage
}

type requestRoomInfoParams []struct {
	inBody wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo
	msg    wire.SNACMessage
	err    error
}

type chatNavParams struct {
	createRoomParams
	exchangeInfoParams
	requestChatRightsParams
	requestRoomInfoParams
}

type channelMsgToHostParamsChat []struct {
	sender state.IdentScreenName
	inBody wire.SNAC_0x0E_0x05_ChatChannelMsgToHost
	result *wire.SNACMessage
	err    error
}

type channelMsgToHostParamsICBM []struct {
	sender  state.IdentScreenName
	inFrame wire.SNACFrame
	inBody  wire.SNAC_0x04_0x06_ICBMChannelMsgToHost
	result  *wire.SNACMessage
	err     error
}

type clientEventParams []struct {
	sess    *state.Session
	inFrame wire.SNACFrame
	inBody  wire.SNAC_0x04_0x14_ICBMClientEvent
	err     error
}

type evilRequestParams []struct {
	me     state.IdentScreenName
	inBody wire.SNAC_0x04_0x08_ICBMEvilRequest
	msg    wire.SNACMessage
	err    error
}

type parameterQueryParams []struct {
	inFrame wire.SNACFrame
	msg     wire.SNACMessage
}

type clientErrParams []struct {
	sess  *state.Session
	frame wire.SNACFrame
	body  wire.SNAC_0x04_0x0B_ICBMClientErr
	err   error
}

type icbmParams struct {
	channelMsgToHostParamsICBM
	clientEventParams
	evilRequestParams
	parameterQueryParams
	clientErrParams
}

type clientOnlineParams []struct {
	body wire.SNAC_0x01_0x02_OServiceClientOnline
	me   state.IdentScreenName
	err  error
}

type idleNotificationParams []struct {
	me     state.IdentScreenName
	bodyIn wire.SNAC_0x01_0x11_OServiceIdleNotification
	err    error
}

type serviceRequestParams []struct {
	me     state.IdentScreenName
	bodyIn wire.SNAC_0x01_0x04_OServiceServiceRequest
	msg    wire.SNACMessage
	err    error
}

type oServiceParams struct {
	clientOnlineParams
	idleNotificationParams
	serviceRequestParams
}

type bucpChallengeParams []struct {
	bodyIn  wire.SNAC_0x17_0x06_BUCPChallengeRequest
	newUUID func() uuid.UUID
	msg     wire.SNACMessage
	err     error
}

type bucpLoginParams []struct {
	bodyIn    wire.SNAC_0x17_0x02_BUCPLoginRequest
	newUserFn func(screenName state.DisplayScreenName) (state.User, error)
	msg       wire.SNACMessage
	err       error
}

type flapLoginParams []struct {
	frame     wire.FLAPSignonFrame
	newUserFn func(screenName state.DisplayScreenName) (state.User, error)
	tlv       wire.TLVRestBlock
	err       error
}

type registerBOSSessionParams []struct {
	authCookie []byte
	sess       *state.Session
	err        error
}

type retrieveBOSSessionParams []struct {
	authCookie []byte
	sess       *state.Session
	err        error
}

type registerChatSessionParams []struct {
	authCookie []byte
	sess       *state.Session
	err        error
}

type signoutParams []struct {
	me state.IdentScreenName
}

type signoutChatParams []struct {
	me state.IdentScreenName
}

type authParams struct {
	bucpChallengeParams
	bucpLoginParams
	flapLoginParams
	registerBOSSessionParams
	retrieveBOSSessionParams
	registerChatSessionParams
	signoutParams
	signoutChatParams
}

type setDirInfoParams []struct {
	me     state.IdentScreenName
	inBody wire.SNAC_0x02_0x09_LocateSetDirInfo
	msg    wire.SNACMessage
	err    error
}

type setInfoParams []struct {
	me     state.IdentScreenName
	inBody wire.SNAC_0x02_0x04_LocateSetInfo
	err    error
}

type userInfoQueryParams []struct {
	me     state.IdentScreenName
	inBody wire.SNAC_0x02_0x05_LocateUserInfoQuery
	msg    wire.SNACMessage
	err    error
}

type dirInfoParams []struct {
	body wire.SNAC_0x02_0x0B_LocateGetDirInfo
	msg  wire.SNACMessage
	err  error
}

type locateParams struct {
	setDirInfoParams
	setInfoParams
	userInfoQueryParams
	dirInfoParams
}

type infoQueryParams []struct {
	inBody wire.SNAC_0x0F_0x02_InfoQuery
	msg    wire.SNACMessage
	err    error
}

type dirSearchParams struct {
	infoQueryParams
}

type addDenyListEntriesParams []struct {
	me   state.IdentScreenName
	body wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries
	err  error
}

type addPermListEntriesParams []struct {
	me   state.IdentScreenName
	body wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries
	err  error
}

type delDenyListEntriesParams []struct {
	me   state.IdentScreenName
	body wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries
	err  error
}

type delPermListEntriesParams []struct {
	me   state.IdentScreenName
	body wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries
	err  error
}

type rightsQueryParamsPermitDeny []struct {
	frame wire.SNACFrame
	msg   wire.SNACMessage
}

type permitDenyParams struct {
	addDenyListEntriesParams
	addPermListEntriesParams
	delDenyListEntriesParams
	delPermListEntriesParams
	rightsQueryParamsPermitDeny
}

type registerBuddyListParams []struct {
	user state.IdentScreenName
	err  error
}

type unregisterBuddyListParams []struct {
	user state.IdentScreenName
	err  error
}

type buddyListRegistryParams struct {
	registerBuddyListParams
	unregisterBuddyListParams
}

type setTOCConfigParams []struct {
	user   state.IdentScreenName
	config string
	err    error
}

type userParams []struct {
	screenName   state.IdentScreenName
	returnedUser *state.User
	err          error
}

type tocConfigParams struct {
	setTOCConfigParams
	userParams
}

type mockParams struct {
	authParams
	buddyListRegistryParams
	buddyParams
	chatNavParams
	chatParams
	cookieBakerParams
	dirSearchParams
	icbmParams
	locateParams
	oServiceBOSParams  oServiceParams
	oServiceChatParams oServiceParams
	permitDenyParams
	tocConfigParams
}

// crackParams holds multiple scenarios for the Crack method.
type crackParams []struct {
	data       []byte
	returnData []byte
	returnErr  error
}

// issueParams holds multiple scenarios for the Issue method.
type issueParams []struct {
	data       []byte
	returnData []byte
	returnErr  error
}

// cookieBakerParams groups the method scenarios for a CookieBaker.
type cookieBakerParams struct {
	crackParams crackParams
	issueParams issueParams
}

// matchSession matches a mock call based session ident screen name.
func matchSession(mustMatch state.IdentScreenName) interface{} {
	return mock.MatchedBy(func(s *state.Session) bool {
		return mustMatch == s.IdentScreenName()
	})
}

// newTestSession creates a session object with 0 or more functional options
// applied
func newTestSession(screenName state.DisplayScreenName, options ...func(session *state.Session)) *state.Session {
	s := state.NewSession()
	s.SetIdentScreenName(screenName.IdentScreenName())
	s.SetDisplayScreenName(screenName)
	for _, op := range options {
		op(s)
	}
	return s
}
