package handler

import (
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

// mockParams is a helper struct that centralizes mock function call parameters
// in one place for a table test
type mockParams struct {
	feedbagManagerParams
	messageRelayerParams
	profileManagerParams
}

// feedbagManagerParams is a helper struct that contains mock parameters for
// FeedbagManager methods
type feedbagManagerParams struct {
	blockedParams
	interestedUsersParams
	upsertParams
	buddiesParams
	retrieveParams
	lastModifiedParams
	deleteParams
}

// blockedParams is the list of parameters passed at the mock
// FeedbagManager.Blocked call site
type blockedParams []struct {
	screenName1 string
	screenName2 string
	result      state.BlockedState
	err         error
}

// interestedUsersParams is the list of parameters passed at the mock
// FeedbagManager.InterestedUsers call site
type interestedUsersParams []struct {
	screenName string
	users      []string
}

// upsertParams is the list of parameters passed at the mock
// FeedbagManager.Upsert call site
type upsertParams []struct {
	screenName string
	items      []oscar.FeedbagItem
}

// buddiesParams is the list of parameters passed at the mock
// FeedbagManager.Buddies call site
type buddiesParams []struct {
	screenName string
	results    []string
}

// retrieveParams is the list of parameters passed at the mock
// FeedbagManager.Retrieve call site
type retrieveParams []struct {
	screenName string
	results    []oscar.FeedbagItem
}

// lastModifiedParams is the list of parameters passed at the mock
// FeedbagManager.LastModified call site
type lastModifiedParams []struct {
	screenName string
	result     time.Time
}

// deleteParams is the list of parameters passed at the mock
// FeedbagManager.Delete call site
type deleteParams []struct {
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
// MessageRelayer.BroadcastToScreenNames call site
type broadcastToScreenNamesParams []struct {
	screenNames []string
	message     oscar.SNACMessage
}

// sendToScreenNameParams is the list of parameters passed at the mock
// MessageRelayer.SendToScreenName call site
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

// sessOptWarning sets a warning level on the session object
func sessOptWarning(level uint16) func(session *state.Session) {
	return func(session *state.Session) {
		session.IncreaseWarning(level)
	}
}

// sessOptCannedID sets a canned session ID ("user-sess-id") on the session
// object
func sessOptCannedID(session *state.Session) {
	session.SetID("user-sess-id")
}

// sessOptCannedID sets a canned session ID ("user-sess-id") on the session
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
