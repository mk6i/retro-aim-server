package state

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
)

var CapChat, _ = uuid.MustParse("748F2420-6287-11D1-8222-444553540000").MarshalBinary()

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

type Session struct {
	awayMessage string
	closed      bool
	id          string
	idle        bool
	idleTime    time.Time
	invisible   bool
	msgCh       chan oscar.XMessage
	mutex       sync.RWMutex
	screenName  string
	signonTime  time.Time
	stopCh      chan struct{}
	warning     uint16
}

func (s *Session) IncreaseWarning(incr uint16) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.warning += incr
}

func (s *Session) SetInvisible(invisible bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.invisible = invisible
}

func (s *Session) SetScreenName(screenName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.screenName = screenName
}

func (s *Session) ScreenName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.screenName
}

func (s *Session) SetID(ID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.id = ID
}

func (s *Session) SetSignonTime(t time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.signonTime = t
}

func (s *Session) ID() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.id
}

func (s *Session) Invisible() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.invisible
}

func (s *Session) SetIdle(dur time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.idle = true
	// set the time the user became idle
	s.idleTime = time.Now().Add(-dur)
}

func (s *Session) SetActive() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.idle = false
}

func (s *Session) Idle() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.idle
}

func (s *Session) SetAwayMessage(awayMessage string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.awayMessage = awayMessage
}

func (s *Session) AwayMessage() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.awayMessage
}

func (s *Session) TLVUserInfo() oscar.TLVUserInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return oscar.TLVUserInfo{
		ScreenName:   s.screenName,
		WarningLevel: s.warning,
		TLVBlock: oscar.TLVBlock{
			TLVList: s.UserInfo(),
		},
	}
}

func (s *Session) UserInfo() oscar.TLVList {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// sign-in timestamp
	tlvs := oscar.TLVList{}

	tlvs.AddTLV(oscar.NewTLV(0x03, uint32(s.signonTime.Unix())))

	// away message status
	if s.awayMessage != "" {
		tlvs.AddTLV(oscar.NewTLV(0x01, uint16(0x0010)|uint16(0x0020)))
	} else {
		tlvs.AddTLV(oscar.NewTLV(0x01, uint16(0x0010)))
	}

	// invisibility status
	if s.invisible {
		tlvs.AddTLV(oscar.NewTLV(0x06, uint16(0x0100)))
	} else {
		tlvs.AddTLV(oscar.NewTLV(0x06, uint16(0x0000)))
	}

	// idle status
	if s.idle {
		tlvs.AddTLV(oscar.NewTLV(0x04, uint16(time.Now().Sub(s.idleTime).Seconds())))
	} else {
		tlvs.AddTLV(oscar.NewTLV(0x04, uint16(0)))
	}

	// capabilities
	var caps []byte
	// chat capability
	caps = append(caps, CapChat...)
	tlvs.AddTLV(oscar.NewTLV(0x0D, caps))

	return tlvs
}

func (s *Session) Warning() uint16 {
	var w uint16
	s.mutex.RLock()
	w = s.warning
	s.mutex.RUnlock()
	return w
}

func (s *Session) RecvMessage() chan oscar.XMessage {
	return s.msgCh
}

func (s *Session) SendMessage(msg oscar.XMessage) SessSendStatus {
	s.mutex.Lock()
	if s.closed {
		return SessSendClosed
	}
	s.mutex.Unlock()
	select {
	case s.msgCh <- msg:
		return SessSendOK
	case <-s.stopCh:
		return SessSendClosed
	default:
		return SessQueueFull
	}
}

func (s *Session) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return
	}
	close(s.stopCh)
	s.closed = true
}

func (s *Session) Closed() <-chan struct{} {
	return s.stopCh
}

func NewSession() *Session {
	return &Session{
		msgCh:      make(chan oscar.XMessage, 1000),
		stopCh:     make(chan struct{}),
		signonTime: time.Now(),
	}
}
