package state

import (
	"sync"
	"time"

	"github.com/mk6i/retro-aim-server/wire"
)

// SessSendStatus is the result of sending a message to a user.
type SessSendStatus int

const (
	// SessSendOK indicates message was sent to recipient
	SessSendOK SessSendStatus = iota
	// SessSendClosed indicates send did not complete because session is closed
	SessSendClosed
	// SessQueueFull indicates send failed due to full queue -- client is likely
	// dead
	SessQueueFull
)

// Session represents a user's current session. Unless stated otherwise, all
// methods may be safely accessed by multiple goroutines.
type Session struct {
	awayMessage    string
	chatRoomCookie string
	closed         bool
	id             string
	idle           bool
	idleTime       time.Time
	invisible      bool
	msgCh          chan wire.SNACMessage
	mutex          sync.RWMutex
	nowFn          func() time.Time
	signonComplete bool
	screenName     string
	signonTime     time.Time
	stopCh         chan struct{}
	warning        uint16
	caps           [][16]byte
}

// NewSession returns a new instance of Session. By default, the user may have
// up to 1000 pending messages before blocking.
func NewSession() *Session {
	return &Session{
		msgCh:      make(chan wire.SNACMessage, 1000),
		nowFn:      time.Now,
		stopCh:     make(chan struct{}),
		signonTime: time.Now(),
		caps:       make([][16]byte, 0),
	}
}

// IncrementWarning increments the user's warning level. To decrease, pass a
// negative increment value.
func (s *Session) IncrementWarning(incr uint16) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.warning += incr
}

// SetInvisible toggles the user's invisibility status.
func (s *Session) SetInvisible(invisible bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.invisible = invisible
}

// Invisible returns true if the user is idle.
func (s *Session) Invisible() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.invisible
}

// SetScreenName sets the user's screen name.
func (s *Session) SetScreenName(screenName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.screenName = screenName
}

// ScreenName returns the user's screen name.
func (s *Session) ScreenName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.screenName
}

// SetID sets the user's session ID.
func (s *Session) SetID(ID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.id = ID
}

// ID returns the user's session ID.
func (s *Session) ID() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.id
}

// SetSignonTime sets the user's sign-ontime.
func (s *Session) SetSignonTime(t time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.signonTime = t
}

// SetIdle sets the user's idle state.
func (s *Session) SetIdle(dur time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.idle = true
	// set the time the user became idle
	s.idleTime = s.nowFn().Add(-dur)
}

// UnsetIdle removes the user's idle state.
func (s *Session) UnsetIdle() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.idle = false
}

// SetAwayMessage sets the user's away message.
func (s *Session) SetAwayMessage(awayMessage string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.awayMessage = awayMessage
}

// AwayMessage returns the user's away message.
func (s *Session) AwayMessage() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.awayMessage
}

// SetChatRoomCookie sets the chatRoomCookie for the chat room the user is currently in.
func (s *Session) SetChatRoomCookie(cookie string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.chatRoomCookie = cookie
}

// ChatRoomCookie gets the chatRoomCookie for the chat room the user is currently in.
func (s *Session) ChatRoomCookie() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.chatRoomCookie
}

// SignonComplete indicates whether the client has completed the sign-on sequence.
func (s *Session) SignonComplete() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.signonComplete
}

// SetSignonComplete indicates that the client has completed the sign-on sequence.
func (s *Session) SetSignonComplete() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.signonComplete = true
}

// TLVUserInfo returns a TLV list containing session information required by
// multiple SNAC message types that convey user information.
func (s *Session) TLVUserInfo() wire.TLVUserInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return wire.TLVUserInfo{
		ScreenName:   s.screenName,
		WarningLevel: s.warning,
		TLVBlock: wire.TLVBlock{
			TLVList: s.userInfo(),
		},
	}
}

func (s *Session) userInfo() wire.TLVList {
	tlvs := wire.TLVList{}

	// sign-in timestamp
	tlvs.Append(wire.NewTLV(wire.OServiceUserInfoSignonTOD, uint32(s.signonTime.Unix())))

	// away message status
	if s.awayMessage != "" {
		tlvs.Append(wire.NewTLV(wire.OServiceUserInfoUserFlags, wire.OServiceUserFlagOSCARFree|wire.OServiceUserFlagUnavailable))
	} else {
		tlvs.Append(wire.NewTLV(wire.OServiceUserInfoUserFlags, wire.OServiceUserFlagOSCARFree))
	}

	// reflects invisibility toggle status back to toggling client
	if s.invisible {
		tlvs.Append(wire.NewTLV(wire.OServiceUserInfoStatus, wire.OServiceUserStatusInvisible))
	} else {
		tlvs.Append(wire.NewTLV(wire.OServiceUserInfoStatus, uint32(0)))
	}

	// idle status
	if s.idle {
		tlvs.Append(wire.NewTLV(wire.OServiceUserInfoIdleTime, uint16(s.nowFn().Sub(s.idleTime).Minutes())))
	}

	// capabilities (buddy icon, chat, etc...)
	if len(s.caps) > 0 {
		tlvs.Append(wire.NewTLV(wire.OServiceUserInfoOscarCaps, s.caps))
	}

	return tlvs
}

// SetCaps sets capability UUIDs that represent the features the client
// supports. If set, capability metadata appears in the user info TLV list.
func (s *Session) SetCaps(caps [][16]byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.caps = caps
}

// Caps retrieves user capabilities.
func (s *Session) Caps() [][16]byte {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.caps
}

func (s *Session) Warning() uint16 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var w uint16
	w = s.warning
	return w
}

// ReceiveMessage returns a channel of messages relayed via this session. It
// may only be read by one consumer. The channel never closes; call this method
// in a select block along with Closed in order to detect session closure.
func (s *Session) ReceiveMessage() chan wire.SNACMessage {
	return s.msgCh
}

// RelayMessage receives a SNAC message from a user and passes it on
// asynchronously to the consumer of this session's messages. It returns
// SessSendStatus to indicate whether the message was successfully sent or
// not. This method is non-blocking.
func (s *Session) RelayMessage(msg wire.SNACMessage) SessSendStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.closed {
		return SessSendClosed
	}
	select {
	case s.msgCh <- msg:
		return SessSendOK
	case <-s.stopCh:
		return SessSendClosed
	default:
		return SessQueueFull
	}
}

// Close shuts down the session's ability to relay messages. Once invoked,
// RelayMessage returns SessQueueFull and Closed returns a closed channel.
// It is not possible to re-open message relaying once closed. It is safe to
// call from multiple go routines.
func (s *Session) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return
	}
	close(s.stopCh)
	s.closed = true
}

// Closed blocks until the session is closed.
func (s *Session) Closed() <-chan struct{} {
	return s.stopCh
}
