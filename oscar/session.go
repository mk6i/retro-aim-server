package oscar

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

var errSessNotFound = errors.New("session was not found")

type Session struct {
	ID          string
	ScreenName  string
	msgCh       chan *XMessage
	stopCh      chan struct{}
	Mutex       sync.RWMutex
	Warning     uint16
	AwayMessage string
	SignonTime  time.Time
	invisible   bool
	idle        bool
	idleTime    time.Time
}

func (s *Session) IncreaseWarning(incr uint16) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.Warning += incr
}

func (s *Session) SetInvisible(invisible bool) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.invisible = invisible
}

func (s *Session) Invisible() bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.invisible
}

func (s *Session) SetIdle(dur time.Duration) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.idle = true
	// set the time the user became idle
	s.idleTime = time.Now().Add(-dur)
}

func (s *Session) SetActive() {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.idle = false
}

func (s *Session) Idle() bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.idle
}

func (s *Session) SetAwayMessage(awayMessage string) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.AwayMessage = awayMessage
}

func (s *Session) GetAwayMessage() string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.AwayMessage
}

func (s *Session) GetUserInfo() []*TLV {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	// sign-in timestamp
	tlvs := []*TLV{
		{
			tType: 0x03,
			val:   uint32(s.SignonTime.Unix()),
		},
	}

	// away message status
	userFlags := &TLV{
		tType: 0x01,
		val:   uint16(0x0010), // AIM client
	}
	if s.AwayMessage != "" {
		userFlags.val = userFlags.val.(uint16) | uint16(0x0020)
	}
	tlvs = append(tlvs, userFlags)

	// invisibility status
	status := &TLV{
		tType: 0x06,
		val:   uint16(0x0000),
	}
	if s.invisible {
		status.val = status.val.(uint16) | uint16(0x0100)
	}
	tlvs = append(tlvs, status)

	// idle status
	idle := &TLV{
		tType: 0x04,
		val:   uint16(0),
	}
	if s.idle {
		idle.val = uint16(time.Now().Sub(s.idleTime).Seconds())
	}
	tlvs = append(tlvs, idle)

	// capabilities
	caps := &TLV{
		tType: 0x0D,
		val:   []byte{},
	}
	// chat capability
	caps.val = append(caps.val.([]byte), CAP_CHAT...)
	tlvs = append(tlvs, caps)

	return tlvs
}

func (s *Session) GetWarning() uint16 {
	var w uint16
	s.Mutex.RLock()
	w = s.Warning
	s.Mutex.RUnlock()
	return w
}

func (s *Session) RecvMessage() chan *XMessage {
	return s.msgCh
}

func (s *Session) SendMessage(msg *XMessage) {
	select {
	case <-s.stopCh:
		return
	case s.msgCh <- msg:
	}
}

func (s *Session) Close() {
	close(s.stopCh)
}

type SessionManager struct {
	store    map[string]*Session
	mapMutex sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		store: make(map[string]*Session),
	}
}

func (s *SessionManager) Retrieve(ID string) (*Session, bool) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	sess, found := s.store[ID]
	return sess, found
}

func (s *SessionManager) RetrieveByScreenName(screenName string) (*Session, error) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if screenName == sess.ScreenName {
			return sess, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", errSessNotFound, screenName)
}

func (s *SessionManager) RetrieveByScreenNames(screenNames []string) []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var ret []*Session
	for _, sn := range screenNames {
		for _, sess := range s.store {
			if sn == sess.ScreenName {
				ret = append(ret, sess)
			}
		}
	}
	return ret
}

func (s *SessionManager) NewSession() (*Session, error) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	sess := &Session{
		ID:         id.String(),
		msgCh:      make(chan *XMessage, 1),
		stopCh:     make(chan struct{}),
		SignonTime: time.Now(),
	}
	s.store[sess.ID] = sess
	return sess, nil
}

func (s *SessionManager) Remove(sess *Session) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	delete(s.store, sess.ID)
}
