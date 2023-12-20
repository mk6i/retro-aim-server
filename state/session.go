package state

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
)

// capChat is a UID that indicates a client supports the chat capability
var capChat, _ = uuid.MustParse("748F2420-6287-11D1-8222-444553540000").MarshalBinary()

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
	awayMessage string
	closed      bool
	id          string
	idle        bool
	idleTime    time.Time
	invisible   bool
	msgCh       chan oscar.SNACMessage
	mutex       sync.RWMutex
	nowFn       func() time.Time
	screenName  string
	signonTime  time.Time
	stopCh      chan struct{}
	warning     uint16
}

// NewSession returns a new instance of Session. By default, the user may have
// up to 1000 pending messages before blocking.
func NewSession() *Session {
	return &Session{
		msgCh:      make(chan oscar.SNACMessage, 1000),
		nowFn:      time.Now,
		stopCh:     make(chan struct{}),
		signonTime: time.Now(),
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

// TLVUserInfo returns a TLV list containing session information required by
// multiple SNAC message types that convey user information.
func (s *Session) TLVUserInfo() oscar.TLVUserInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return oscar.TLVUserInfo{
		ScreenName:   s.screenName,
		WarningLevel: s.warning,
		TLVBlock: oscar.TLVBlock{
			TLVList: s.userInfo(),
		},
	}
}

func (s *Session) userInfo() oscar.TLVList {
	// sign-in timestamp
	tlvs := oscar.TLVList{}

	tlvs.AddTLV(oscar.NewTLV(oscar.OServiceUserInfoSignonTOD, uint32(s.signonTime.Unix())))

	// away message status
	if s.awayMessage != "" {
		tlvs.AddTLV(oscar.NewTLV(oscar.OServiceUserInfoUserFlags, oscar.OServiceUserFlagOSCARFree|oscar.OServiceUserFlagUnavailable))
	} else {
		tlvs.AddTLV(oscar.NewTLV(oscar.OServiceUserInfoUserFlags, oscar.OServiceUserFlagOSCARFree))
	}

	// invisibility status
	if s.invisible {
		tlvs.AddTLV(oscar.NewTLV(oscar.OServiceUserInfoStatus, oscar.OServiceUserFlagInvisible))
	} else {
		tlvs.AddTLV(oscar.NewTLV(oscar.OServiceUserInfoStatus, uint16(0x0000)))
	}

	// idle status
	if s.idle {
		tlvs.AddTLV(oscar.NewTLV(oscar.OServiceUserInfoIdleTime, uint16(s.nowFn().Sub(s.idleTime).Seconds())))
	} else {
		tlvs.AddTLV(oscar.NewTLV(oscar.OServiceUserInfoIdleTime, uint16(0)))
	}

	// capabilities
	var caps []byte
	// chat capability
	caps = append(caps, capChat...)
	tlvs.AddTLV(oscar.NewTLV(oscar.OServiceUserInfoOscarCaps, caps))

	return tlvs
}

// Warning returns the user's warning level.
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
func (s *Session) ReceiveMessage() chan oscar.SNACMessage {
	return s.msgCh
}

// RelayMessage receives a SNAC message from a user and passes it on
// asynchronously to the consumer of this session's messages. It returns
// SessSendStatus to indicate whether the message was successfully sent or
// not. This method is non-blocking.
func (s *Session) RelayMessage(msg oscar.SNACMessage) SessSendStatus {
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
