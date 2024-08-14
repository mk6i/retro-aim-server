package foodgroup

import (
	"net/mail"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// mockParams is a helper struct that centralizes mock function call parameters
// in one place for a table test
type mockParams struct {
	accountManagerParams
	bartManagerParams
	buddyBroadcasterParams
	chatMessageRelayerParams
	chatRoomRegistryParams
	cookieBakerParams
	feedbagManagerParams
	icqUserFinderParams
	icqUserUpdaterParams
	legacyBuddyListManagerParams
	messageRelayerParams
	profileManagerParams
	sessionManagerParams
	sessionRetrieverParams
	userManagerParams
}

// sessionRetrieverParams is a helper struct that contains mock parameters for
// SessionRetriever methods
type sessionRetrieverParams struct {
	retrieveSessionParams
}

// retrieveSessionParams is a helper struct that contains mock parameters for
// SessionRetriever methods
type retrieveSessionParams []struct {
	screenName state.IdentScreenName
	result     *state.Session
}

// icqUserFinderParams is a helper struct that contains mock parameters for
// ICQUserFinder methods
type icqUserFinderParams struct {
	findByUINParams
	findByEmailParams
	findByDetailsParams
	findByInterestsParams
}

// findByUINParams is the list of parameters passed at the mock
// ICQUserFinder.FindByUIN call site
type findByUINParams []struct {
	UIN    uint32
	result state.User
	err    error
}

// findByEmailParams is the list of parameters passed at the mock
// ICQUserFinder.FindByEmail call site
type findByEmailParams []struct {
	email  string
	result state.User
	err    error
}

// setBasicInfoParams is the list of parameters passed at the mock
// ICQUserFinder.FindByDetails call site
type findByDetailsParams []struct {
	firstName string
	lastName  string
	nickName  string
	result    []state.User
	err       error
}

// setBasicInfoParams is the list of parameters passed at the mock
// ICQUserFinder.FindByInterests call site
type findByInterestsParams []struct {
	code     uint16
	keywords []string
	result   []state.User
	err      error
}

// icqUserUpdaterParams is a helper struct that contains mock parameters for
// ICQUserUpdater methods
type icqUserUpdaterParams struct {
	setAffiliationsParams
	setBasicInfoParams
	setInterestsParams
	setMoreInfoParams
	setUserNotesParams
	setWorkInfoParams
}

// setAffiliationsParams is the list of parameters passed at the mock
// ICQUserUpdater.SetAffiliations call site
type setAffiliationsParams []struct {
	name state.IdentScreenName
	data state.ICQAffiliations
	err  error
}

// setInterestsParams is the list of parameters passed at the mock
// ICQUserUpdater.SetInterests call site
type setInterestsParams []struct {
	name state.IdentScreenName
	data state.ICQInterests
	err  error
}

// setUserNotesParams is the list of parameters passed at the mock
// ICQUserUpdater.SetUserNotes call site
type setUserNotesParams []struct {
	name state.IdentScreenName
	data state.ICQUserNotes
	err  error
}

// setBasicInfoParams is the list of parameters passed at the mock
// ICQUserUpdater.SetBasicInfo call site
type setBasicInfoParams []struct {
	name state.IdentScreenName
	data state.ICQBasicInfo
	err  error
}

// setWorkInfoParams is the list of parameters passed at the mock
// ICQUserUpdater.SetWorkInfo call site
type setWorkInfoParams []struct {
	name state.IdentScreenName
	data state.ICQWorkInfo
	err  error
}

// setMoreInfoParams is the list of parameters passed at the mock
// ICQUserUpdater.SetMoreInfo call site
type setMoreInfoParams []struct {
	name state.IdentScreenName
	data state.ICQMoreInfo
	err  error
}

// bartManagerParams is a helper struct that contains mock parameters for
// BARTManager methods
type bartManagerParams struct {
	bartManagerRetrieveParams
	bartManagerUpsertParams
}

// bartManagerRetrieveParams is the list of parameters passed at the mock
// BARTManager.BARTRetrieve call site
type bartManagerRetrieveParams []struct {
	itemHash []byte
	result   []byte
}

// bartManagerUpsertParams is the list of parameters passed at the mock
// BARTManager.BARTUpsert call site
type bartManagerUpsertParams []struct {
	itemHash []byte
	payload  []byte
}

// userManagerParams is a helper struct that contains mock parameters for
// UserManager methods
type userManagerParams struct {
	getUserParams
	insertUserParams
}

// getUserParams is the list of parameters passed at the mock
// UserManager.User call site
type getUserParams []struct {
	screenName state.IdentScreenName
	result     *state.User
	err        error
}

// insertUserParams is the list of parameters passed at the mock
// UserManager.InsertUser call site
type insertUserParams []struct {
	user state.User
	err  error
}

// sessionManagerParams is a helper struct that contains mock parameters for
// SessionManager methods
type sessionManagerParams struct {
	emptyParams
	addSessionParams
	removeSessionParams
}

// addSessionParams is the list of parameters passed at the mock
// SessionManager.AddSession call site
type addSessionParams []struct {
	screenName state.DisplayScreenName
	result     *state.Session
}

// removeSessionParams is the list of parameters passed at the mock
// SessionManager.RemoveSession call site
type removeSessionParams []struct {
	screenName state.IdentScreenName
}

// emptyParams is the list of parameters passed at the mock
// SessionManager.Empty call site
type emptyParams []struct {
	result bool
}

// feedbagManagerParams is a helper struct that contains mock parameters for
// FeedbagManager methods
type feedbagManagerParams struct {
	blockedStateParams
	adjacentUsersParams
	feedbagUpsertParams
	buddiesParams
	feedbagParams
	feedbagLastModifiedParams
	feedbagDeleteParams
}

// blockedStateParams is the list of parameters passed at the mock
// FeedbagManager.BlockedState call site
type blockedStateParams []struct {
	screenName1 state.IdentScreenName
	screenName2 state.IdentScreenName
	result      state.BlockedState
	err         error
}

// adjacentUsersParams is the list of parameters passed at the mock
// FeedbagManager.AdjacentUsers call site
type adjacentUsersParams []struct {
	screenName state.IdentScreenName
	users      []state.IdentScreenName
	err        error
}

// feedbagUpsertParams is the list of parameters passed at the mock
// FeedbagManager.FeedbagUpsert call site
type feedbagUpsertParams []struct {
	screenName state.IdentScreenName
	items      []wire.FeedbagItem
}

// buddiesParams is the list of parameters passed at the mock
// FeedbagManager.Buddies call site
type buddiesParams []struct {
	screenName state.IdentScreenName
	results    []state.IdentScreenName
}

// feedbagParams is the list of parameters passed at the mock
// FeedbagManager.Feedbag call site
type feedbagParams []struct {
	screenName state.IdentScreenName
	results    []wire.FeedbagItem
}

// feedbagLastModifiedParams is the list of parameters passed at the mock
// FeedbagManager.FeedbagLastModified call site
type feedbagLastModifiedParams []struct {
	screenName state.IdentScreenName
	result     time.Time
}

// feedbagDeleteParams is the list of parameters passed at the mock
// FeedbagManager.FeedbagDelete call site
type feedbagDeleteParams []struct {
	screenName state.IdentScreenName
	items      []wire.FeedbagItem
}

// messageRelayerParams is a helper struct that contains mock parameters for
// MessageRelayer methods
type messageRelayerParams struct {
	retrieveByScreenNameParams
	relayToScreenNamesParams
	relayToScreenNameParams
}

// retrieveByScreenNameParams is the list of parameters passed at the mock
// MessageRelayer.RetrieveByScreenName call site
type retrieveByScreenNameParams []struct {
	screenName state.IdentScreenName
	sess       *state.Session
}

// relayToScreenNamesParams is the list of parameters passed at the mock
// MessageRelayer.RelayToScreenNames call site
type relayToScreenNamesParams []struct {
	screenNames []state.IdentScreenName
	message     wire.SNACMessage
}

// relayToScreenNameParams is the list of parameters passed at the mock
// MessageRelayer.RelayToScreenName call site
type relayToScreenNameParams []struct {
	cookie     string
	screenName state.IdentScreenName
	message    wire.SNACMessage
}

// profileManagerParams is a helper struct that contains mock parameters for
// ProfileManager methods
type profileManagerParams struct {
	retrieveProfileParams
	setProfileParams
}

// retrieveByScreenNameParams is the list of parameters passed at the mock
// ProfileManager.Profile call site
type retrieveProfileParams []struct {
	screenName state.IdentScreenName
	result     string
	err        error
}

// setProfileParams is the list of parameters passed at the mock
// ProfileManager.SetProfile call site
type setProfileParams []struct {
	screenName state.IdentScreenName
	body       any
}

// chatMessageRelayerParams is a helper struct that contains mock parameters
// for ChatMessageRelayer methods
type chatMessageRelayerParams struct {
	chatAllSessionsParams
	chatRelayToAllExceptParams
	chatRelayToScreenNameParams
}

// chatAllSessionsParams is the list of parameters passed at the mock
// ChatMessageRelayer.AllSessions call site
type chatAllSessionsParams []struct {
	cookie   string
	sessions []*state.Session
	err      error
}

// chatRelayToAllExceptParams is the list of parameters passed at the mock
// ChatMessageRelayer.RelayToAllExcept call site
type chatRelayToAllExceptParams []struct {
	cookie     string
	screenName state.IdentScreenName
	message    wire.SNACMessage
	err        error
}

// chatRelayToScreenNameParams is the list of parameters passed at the mock
// ChatMessageRelayer.RelayToScreenName call site
type chatRelayToScreenNameParams []struct {
	cookie     string
	screenName state.IdentScreenName
	message    wire.SNACMessage
	err        error
}

// legacyBuddyListManagerParams is a helper struct that contains mock
// parameters for LegacyBuddyListManager methods
type legacyBuddyListManagerParams struct {
	addBuddyParams
	deleteBuddyParams
	deleteUserParams
	legacyBuddiesParams
	whoAddedUserParams
}

// legacyBuddiesParams is the list of parameters passed at the mock
// LegacyBuddyListManager.AddBuddy call site
type addBuddyParams []struct {
	userScreenName  state.IdentScreenName
	buddyScreenName state.IdentScreenName
}

// legacyBuddiesParams is the list of parameters passed at the mock
// LegacyBuddyListManager.DeleteBuddy call site
type deleteBuddyParams []struct {
	userScreenName  state.IdentScreenName
	buddyScreenName state.IdentScreenName
}

// deleteUserParams is the list of parameters passed at the mock
// LegacyBuddyListManager.DeleteUser call site
type deleteUserParams []struct {
	userScreenName state.IdentScreenName
}

// legacyBuddiesParams is the list of parameters passed at the mock
// LegacyBuddyListManager.Buddies call site
type legacyBuddiesParams []struct {
	userScreenName state.IdentScreenName
	result         []state.IdentScreenName
}

// whoAddedUserParams is the list of parameters passed at the mock
// LegacyBuddyListManager.WhoAddedUser call site
type whoAddedUserParams []struct {
	userScreenName state.IdentScreenName
	result         []state.IdentScreenName
}

// cookieBakerParams is a helper struct that contains mock parameters for
// CookieBaker methods
type cookieBakerParams struct {
	cookieCrackParams
	cookieIssueParams
}

// cookieCrackParams is the list of parameters passed at the mock
// CookieBaker.Crack call site
type cookieCrackParams []struct {
	cookieIn []byte
	dataOut  []byte
	err      error
}

// cookieIssueParams is the list of parameters passed at the mock
// CookieBaker.Issue call site
type cookieIssueParams []struct {
	dataIn    []byte
	cookieOut []byte
	err       error
}

// accountManagerParams is a helper struct that contains mock parameters for
// accountManager methods
type accountManagerParams struct {
	accountManagerUpdateDisplayScreenNameParams
	accountManagerUpdateEmailAddressParams
	accountManagerEmailAddressByNameParams
	accountManagerUpdateRegStatusParams
	accountManagerRegStatusByNameParams
	accountManagerUpdateConfirmStatusParams
	accountManagerConfirmStatusByNameParams
}

// accountManagerUpdateDisplayScreenNameParams is the list of parameters passed at the mock
// accountManager.UpdateDisplayScreenName call site
type accountManagerUpdateDisplayScreenNameParams []struct {
	displayScreenName state.DisplayScreenName
	err               error
}

// accountManagerUpdateEmailAddressParams is the list of parameters passed at the mock
// accountManager.UpdateEmailAddress call site
type accountManagerUpdateEmailAddressParams []struct {
	emailAddress *mail.Address
	screenName   state.IdentScreenName
	err          error
}

// accountManagerEmailAddressByNameParams is the list of parameters passed at the mock
// accountManager.EmailAddressByName call site
type accountManagerEmailAddressByNameParams []struct {
	screenName   state.IdentScreenName
	emailAddress *mail.Address
	err          error
}

// accountManagerUpdateRegStatusParams is the list of parameters passed at the mock
// accountManager.UpdateRegStatus call site
type accountManagerUpdateRegStatusParams []struct {
	regStatus  uint16
	screenName state.IdentScreenName
	err        error
}

// accountManagerRegStatusByNameParams is the list of parameters passed at the mock
// accountManager.RegStatusByName call site
type accountManagerRegStatusByNameParams []struct {
	screenName state.IdentScreenName
	regStatus  uint16
	err        error
}

// accountManagerUpdateConfirmStatusParams is the list of parameters passed at the mock
// accountManager.UpdateConfirmStatus call site
type accountManagerUpdateConfirmStatusParams []struct {
	confirmStatus bool
	screenName    state.IdentScreenName
	err           error
}

// accountManagerConfirmStatusByNameParams is the list of parameters passed at the mock
// accountManager.ConfirmStatusByName call site
type accountManagerConfirmStatusByNameParams []struct {
	screenName    state.IdentScreenName
	confirmStatus bool
	err           error
}

// buddyBroadcasterParams is a helper struct that contains mock parameters for
// buddyBroadcaster methods
type buddyBroadcasterParams struct {
	broadcastBuddyArrivedParams
	broadcastBuddyDepartedParams
	unicastBuddyArrivedParams
	unicastBuddyDepartedParams
}

// broadcastBuddyArrivedParams is the list of parameters passed at the mock
// buddyBroadcaster.BroadcastBuddyArrived call site
type broadcastBuddyArrivedParams []struct {
	screenName state.IdentScreenName
	err        error
}

// broadcastBuddyDepartedParams is the list of parameters passed at the mock
// buddyBroadcaster.BroadcastBuddyDeparted call site
type broadcastBuddyDepartedParams []struct {
	screenName state.IdentScreenName
	err        error
}

// unicastBuddyArrivedParams is the list of parameters passed at the mock
// buddyBroadcaster.UnicastBuddyArrived call site
type unicastBuddyArrivedParams []struct {
	from state.IdentScreenName
	to   state.IdentScreenName
	err  error
}

// unicastBuddyDepartedParams is the list of parameters passed at the mock
// buddyBroadcaster.UnicastBuddyDeparted call site
type unicastBuddyDepartedParams []struct {
	from state.IdentScreenName
	to   state.IdentScreenName
	err  error
}

// chatRoomRegistryParams is a helper struct that contains mock parameters for
// ChatRoomRegistry methods
type chatRoomRegistryParams struct {
	chatRoomByCookieParams
	chatRoomByNameParams
	createChatRoomParams
}

// chatRoomByCookieParams is the list of parameters passed at the mock
// ChatRoomRegistry.ChatRoomByCookie call site
type chatRoomByCookieParams []struct {
	cookie string
	room   state.ChatRoom
	err    error
}

// chatRoomByCookieParams is the list of parameters passed at the mock
// ChatRoomRegistry.ChatRoomByName call site
type chatRoomByNameParams []struct {
	exchange uint16
	name     string
	room     state.ChatRoom
	err      error
}

// createChatRoomParams is the list of parameters passed at the mock
// ChatRoomRegistry.CreateChatRoom call site
type createChatRoomParams []struct {
	room *state.ChatRoom
	err  error
}

// sessOptWarning sets a warning level on the session object
func sessOptWarning(level uint16) func(session *state.Session) {
	return func(session *state.Session) {
		session.IncrementWarning(level)
	}
}

// sessOptAwayMessage sets away message on the session object
func sessOptAwayMessage(awayMessage string) func(session *state.Session) {
	return func(session *state.Session) {
		session.SetAwayMessage(awayMessage)
	}
}

// sessOptCannedAwayMessage sets a canned away message ("this is my away
// message!") on the session object
func sessOptCannedAwayMessage(session *state.Session) {
	session.SetAwayMessage("this is my away message!")
}

// sessOptCannedSignonTime sets a canned sign-on time (1696790127565) on the
// session object
func sessOptCannedSignonTime(session *state.Session) {
	session.SetSignonTime(time.UnixMilli(1696790127565))
}

// sessOptChatRoomCookie sets cookie on the session object
func sessOptChatRoomCookie(cookie string) func(session *state.Session) {
	return func(session *state.Session) {
		session.SetChatRoomCookie(cookie)
	}
}

// sessOptInvisible sets the invisible flag to true on the session
// object
func sessOptInvisible(session *state.Session) {
	session.SetUserStatusBitmask(wire.OServiceUserStatusInvisible)
}

// sessOptIdle sets the idle flag to dur on the session object
func sessOptIdle(dur time.Duration) func(session *state.Session) {
	return func(session *state.Session) {
		session.SetIdle(dur)
	}
}

// sessOptSignonComplete sets the sign on complete flag to true
func sessOptSignonComplete(session *state.Session) {
	session.SetSignonComplete()
}

// sessOptCaps sets caps
func sessOptCaps(caps [][16]byte) func(session *state.Session) {
	return func(session *state.Session) {
		session.SetCaps(caps)
	}
}

// sessOptCaps sets caps
func sessOptUIN(UIN uint32) func(session *state.Session) {
	return func(session *state.Session) {
		session.SetUIN(UIN)
	}
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

func userInfoWithBARTIcon(sess *state.Session, bid wire.BARTID) wire.TLVUserInfo {
	info := sess.TLVUserInfo()
	info.Append(wire.NewTLV(wire.OServiceUserInfoBARTInfo, bid))
	return info
}

// matchSession matches a mock call based session ident screen name.
func matchSession(mustMatch state.IdentScreenName) interface{} {
	return mock.MatchedBy(func(s *state.Session) bool {
		return mustMatch == s.IdentScreenName()
	})
}
