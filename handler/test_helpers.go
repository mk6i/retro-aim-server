package handler

import (
	"time"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

// mockParams is a helper struct that centralizes mock function call parameters
// in one place for a table test
type mockParams struct {
	chatMessageRelayerParams
	chatRegistryParams
	feedbagManagerParams
	messageRelayerParams
	profileManagerParams
	sessionManagerParams
	userManagerParams
}

type chatRegistryParams struct {
	chatRegistryRetrieveParams
}

type chatRegistryRetrieveParams struct {
	chatID         string
	retChatRoom    state.ChatRoom
	retChatSessMgr any
	err            error
}

// userManagerParams is a helper struct that contains mock parameters for
// UserManager methods
type userManagerParams struct {
	getUserParams
	upsertUserParams
}

// getUserParams is the list of parameters passed at the mock
// UserManager.GetUser call site
type getUserParams []struct {
	screenName string
	result     *state.User
	err        error
}

// upsertUserParams is the list of parameters passed at the mock
// UserManager.UpsertUser call site
type upsertUserParams []struct {
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
	sessID     string
	screenName string
	result     *state.Session
}

// removeSessionParams is the list of parameters passed at the mock
// SessionManager.RemoveSession call site
type removeSessionParams []struct {
	sess *state.Session
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
	interestedUsersParams
	feedbagUpsertParams
	buddiesParams
	feedbagParams
	feedbagLastModifiedParams
	feedbagDeleteParams
}

// blockedStateParams is the list of parameters passed at the mock
// FeedbagManager.BlockedState call site
type blockedStateParams []struct {
	screenName1 string
	screenName2 string
	result      state.BlockedState
	err         error
}

// interestedUsersParams is the list of parameters passed at the mock
// FeedbagManager.AdjacentUsers call site
type interestedUsersParams []struct {
	screenName string
	users      []string
	err        error
}

// feedbagUpsertParams is the list of parameters passed at the mock
// FeedbagManager.FeedbagUpsert call site
type feedbagUpsertParams []struct {
	screenName string
	items      []oscar.FeedbagItem
}

// buddiesParams is the list of parameters passed at the mock
// FeedbagManager.Buddies call site
type buddiesParams []struct {
	screenName string
	results    []string
}

// feedbagParams is the list of parameters passed at the mock
// FeedbagManager.Feedbag call site
type feedbagParams []struct {
	screenName string
	results    []oscar.FeedbagItem
}

// feedbagLastModifiedParams is the list of parameters passed at the mock
// FeedbagManager.FeedbagLastModified call site
type feedbagLastModifiedParams []struct {
	screenName string
	result     time.Time
}

// feedbagDeleteParams is the list of parameters passed at the mock
// FeedbagManager.FeedbagDelete call site
type feedbagDeleteParams []struct {
	screenName string
	items      []oscar.FeedbagItem
}

// messageRelayerParams is a helper struct that contains mock parameters for
// MessageRelayer methods
type messageRelayerParams struct {
	retrieveByScreenNameParams
	broadcastToScreenNamesParams
	sendToScreenNameParams
}

// retrieveByScreenNameParams is the list of parameters passed at the mock
// MessageRelayer.RetrieveByScreenName call site
type retrieveByScreenNameParams []struct {
	screenName string
	sess       *state.Session
}

// broadcastToScreenNamesParams is the list of parameters passed at the mock
// MessageRelayer.RelayToScreenNames call site
type broadcastToScreenNamesParams []struct {
	screenNames []string
	message     oscar.SNACMessage
}

// sendToScreenNameParams is the list of parameters passed at the mock
// MessageRelayer.RelayToScreenName call site
type sendToScreenNameParams []struct {
	screenName string
	message    oscar.SNACMessage
}

// profileManagerParams is a helper struct that contains mock parameters for
// ProfileManager methods
type profileManagerParams struct {
	retrieveProfileParams
	upsertProfileParams
}

// retrieveByScreenNameParams is the list of parameters passed at the mock
// ProfileManager.RetrieveProfile call site
type retrieveProfileParams []struct {
	screenName string
	result     string
	err        error
}

// upsertProfileParams is the list of parameters passed at the mock
// ProfileManager.UpsertProfile call site
type upsertProfileParams []struct {
	screenName string
	body       any
}

// chatMessageRelayerParams is a helper struct that contains mock parameters
// for ChatMessageRelayer methods
type chatMessageRelayerParams struct {
	broadcastExceptParams
}

// broadcastExceptParams is the list of parameters passed at the mock
// ChatMessageRelayer.RelayToAllExcept call site
type broadcastExceptParams []struct {
	except  *state.Session
	message oscar.SNACMessage
}

// sessOptWarning sets a warning level on the session object
func sessOptWarning(level uint16) func(session *state.Session) {
	return func(session *state.Session) {
		session.IncrementWarning(level)
	}
}

// sessOptCannedID sets a canned session ID ("user-userSession-id") on the session
// object
func sessOptCannedID(session *state.Session) {
	session.SetID("user-userSession-id")
}

// sessOptCannedID sets a canned session ID ("user-userSession-id") on the session
// object
func sessOptID(ID string) func(session *state.Session) {
	return func(session *state.Session) {
		session.SetID(ID)
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

// sessOptChatID sets chat ID on the session object
func sessOptChatID(chatID string) func(session *state.Session) {
	return func(session *state.Session) {
		session.SetChatID(chatID)
	}
}

// sessOptCannedSignonTime sets the invisible flag to true on the session
// object
func sessOptInvisible(session *state.Session) {
	session.SetInvisible(true)
}

// sessOptIdle sets the idle flag to dur on the session object
func sessOptIdle(dur time.Duration) func(session *state.Session) {
	return func(session *state.Session) {
		session.SetIdle(dur)
	}
}

// newTestSession creates a session object with 0 or more functional options
// applied
func newTestSession(screenName string, options ...func(session *state.Session)) *state.Session {
	s := state.NewSession()
	s.SetScreenName(screenName)
	for _, op := range options {
		op(s)
	}
	return s
}
