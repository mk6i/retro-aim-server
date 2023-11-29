package handler

import (
	"time"

	"github.com/mkaminski/goaim/state"
)

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
