package handler

import (
	"github.com/mkaminski/goaim/server"
	"time"
)

// sessOptWarning sets a warning level on the session object
func sessOptWarning(level uint16) func(session *server.Session) {
	return func(session *server.Session) {
		session.IncreaseWarning(level)
	}
}

// sessOptCannedID sets a canned session ID ("user-sess-id") on the session
// object
func sessOptCannedID(session *server.Session) {
	session.SetID("user-sess-id")
}

// sessOptCannedID sets a canned session ID ("user-sess-id") on the session
// object
func sessOptID(ID string) func(session *server.Session) {
	return func(session *server.Session) {
		session.SetID(ID)
	}
}

// sessOptCannedAwayMessage sets a canned away message ("this is my away
// message!") on the session object
func sessOptCannedAwayMessage(session *server.Session) {
	session.SetAwayMessage("this is my away message!")
}

// sessOptCannedSignonTime sets a canned sign-on time (1696790127565) on the
// session object
func sessOptCannedSignonTime(session *server.Session) {
	session.SetSignonTime(time.UnixMilli(1696790127565))
}

// sessOptCannedSignonTime sets the invisible flag to true on the session
// object
func sessOptInvisible(session *server.Session) {
	session.SetInvisible(true)
}

// newTestSession creates a session object with 0 or more functional options
// applied
func newTestSession(screenName string, options ...func(session *server.Session)) *server.Session {
	s := server.NewSession()
	s.SetScreenName(screenName)
	for _, op := range options {
		op(s)
	}
	return s
}
