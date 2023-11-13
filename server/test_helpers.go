package server

import (
	"github.com/mkaminski/goaim/user"
	"time"
)

// sessOptWarning sets a warning level on the session object
func sessOptWarning(level uint16) func(session *user.Session) {
	return func(session *user.Session) {
		session.IncreaseWarning(level)
	}
}

// sessOptCannedID sets a canned session ID ("user-sess-id") on the session
// object
func sessOptCannedID(session *user.Session) {
	session.SetID("user-sess-id")
}

// sessOptCannedAwayMessage sets a canned away message ("this is my away
// message!") on the session object
func sessOptCannedAwayMessage(session *user.Session) {
	session.SetAwayMessage("this is my away message!")
}

// sessOptCannedSignonTime sets a canned sign-on time (1696790127565) on the
// session object
func sessOptCannedSignonTime(session *user.Session) {
	session.SetSignonTime(time.UnixMilli(1696790127565))
}

// sessOptCannedSignonTime sets the invisible flag to true on the session
// object
func sessOptInvisible(session *user.Session) {
	session.SetInvisible(true)
}

// newTestSession creates a session object with 0 or more functional options
// applied
func newTestSession(screenName string, options ...func(session *user.Session)) *user.Session {
	s := user.NewSession()
	s.SetScreenName(screenName)
	for _, op := range options {
		op(s)
	}
	return s
}
