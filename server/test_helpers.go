package server

import (
	"time"
)

// sessOptWarning sets a warning level on the session object
func sessOptWarning(level uint16) func(session *Session) {
	return func(session *Session) {
		session.IncreaseWarning(level)
	}
}

// sessOptCannedID sets a canned session ID ("user-sess-id") on the session
// object
func sessOptCannedID(session *Session) {
	session.SetID("user-sess-id")
}

// sessOptCannedAwayMessage sets a canned away message ("this is my away
// message!") on the session object
func sessOptCannedAwayMessage(session *Session) {
	session.SetAwayMessage("this is my away message!")
}

// sessOptCannedSignonTime sets a canned sign-on time (1696790127565) on the
// session object
func sessOptCannedSignonTime(session *Session) {
	session.SetSignonTime(time.UnixMilli(1696790127565))
}

// sessOptCannedSignonTime sets the invisible flag to true on the session
// object
func sessOptInvisible(session *Session) {
	session.SetInvisible(true)
}

// newTestSession creates a session object with 0 or more functional options
// applied
func newTestSession(screenName string, options ...func(session *Session)) *Session {
	s := NewSession()
	s.SetScreenName(screenName)
	for _, op := range options {
		op(s)
	}
	return s
}
