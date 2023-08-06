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
}

func (s *Session) IncreaseWarning(incr uint16) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.Warning += incr
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

	tlvs := []*TLV{
		{
			tType: 0x03,
			val:   uint32(s.SignonTime.Unix()),
		},
	}

	if s.AwayMessage != "" {
		tlvs = append(tlvs, &TLV{
			tType: 0x01,
			val:   uint16(0x0010 | 0x0020),
		})
	} else {
		tlvs = append(tlvs, &TLV{
			tType: 0x01,
			val:   uint16(0x0010),
		})

	}

	tlvs = append(tlvs, &TLV{
		tType: 0x06,
		val:   uint16(0x0000),
	})

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
